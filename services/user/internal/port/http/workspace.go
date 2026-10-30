package http

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/nostosx/servex-contrib/auth/login"
	"github.com/nostosx/servex-contrib/auth/password"
	etcdx "github.com/nostosx/servex-contrib/coordination/etcd"
	"github.com/nostosx/servex-contrib/middleware/auth"
	s3x "github.com/nostosx/servex-contrib/storage/s3"
	"github.com/nostosx/servex-contrib/transport/httpx"
	"github.com/nostosx/servex/messaging/jobqueue"
	"github.com/nostosx/servex/storage/cache"

	"github.com/nostosx/servex-examples/application/identity"
	"github.com/nostosx/servex-examples/application/notes"
	notecommand "github.com/nostosx/servex-examples/application/notes/command"
	notequery "github.com/nostosx/servex-examples/application/notes/query"
	"github.com/nostosx/servex-examples/infrastructure/config"
	serviceapp "github.com/nostosx/servex-examples/services/user/internal/application"
)

// Workspace receives ordinary constructor dependencies, never a DI container.
type Workspace struct {
	commands *serviceapp.Command
	queries  *serviceapp.Query
	login    *login.Manager[identity.Principal]
	hasher   *password.Hasher
	config   config.Composition
	cache    cache.Cache
}

func NewWorkspace(commands *serviceapp.Command, queries *serviceapp.Query, manager *login.Manager[identity.Principal], hasher *password.Hasher, cfg config.Composition, cached cache.Cache) *Workspace {
	return &Workspace{commands: commands, queries: queries, login: manager, hasher: hasher, config: cfg, cache: cached}
}

func (h *Workspace) Register(server *httpx.Server, jobs jobqueue.Enqueuer, objects *s3x.Store, coordinator *etcdx.Coordinator) {
	server.HandleHTTP("POST /auth/password", http.HandlerFunc(h.signIn))
	server.HandleHTTP("POST /auth/logout", http.HandlerFunc(h.login.Logout))
	server.HandleHTTP("GET /auth/oidc", http.HandlerFunc(h.login.Begin))
	server.HandleHTTP("GET /auth/callback", http.HandlerFunc(h.login.Callback))
	protected := http.NewServeMux()
	protected.HandleFunc("GET /v1/notes", h.list)
	protected.HandleFunc("POST /v1/notes", h.create)
	protected.HandleFunc("PUT /v1/notes/{id}", h.update)
	protected.HandleFunc("DELETE /v1/notes/{id}", h.delete)
	protected.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
		p, _ := auth.FromContext[identity.Principal](r.Context())
		respond(w, http.StatusOK, p)
	})
	if jobs != nil {
		protected.HandleFunc("POST /v1/jobs/refresh", func(w http.ResponseWriter, r *http.Request) {
			id, err := jobs.Enqueue(r.Context(), &jobqueue.Job{Type: "example.refresh"})
			if err != nil {
				failure(w, err)
				return
			}
			respond(w, http.StatusAccepted, map[string]string{"job_id": id})
		})
	}
	protected.HandleFunc("GET /v1/jobs/last-run", func(w http.ResponseWriter, r *http.Request) {
		value, err := h.cache.Get(r.Context(), "refresh:last-run")
		if errors.Is(err, cache.ErrNotFound) {
			respond(w, http.StatusOK, map[string]string{"at": ""})
			return
		}
		if err != nil {
			failure(w, err)
			return
		}
		respond(w, http.StatusOK, map[string]string{"at": string(value)})
	})
	if objects != nil {
		h.registerObjects(protected, objects)
	}
	if coordinator != nil {
		protected.HandleFunc("POST /v1/coordination/probe", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			lock, err := coordinator.Acquire(ctx, "demo-probe")
			if err != nil {
				failure(w, err)
				return
			}
			// This probe performs no protected external write. Such writes must enforce fencing.
			if err := lock.Release(r.Context()); err != nil {
				failure(w, err)
				return
			}
			respond(w, http.StatusOK, map[string]int64{"revision": lock.Revision})
		})
	}
	server.HandleHTTP("/v1/notes", h.login.Require(protected))
	server.HandleHTTP("/v1/notes/", h.login.Require(protected))
	server.HandleHTTP("/v1/me", h.login.Require(protected))
	server.HandleHTTP("/v1/jobs/", h.login.Require(protected))
	server.HandleHTTP("/v1/objects", h.login.Require(protected))
	server.HandleHTTP("/v1/objects/", h.login.Require(protected))
	server.HandleHTTP("/v1/coordination/", h.login.Require(protected))
}

func (h *Workspace) signIn(w http.ResponseWriter, r *http.Request) {
	if h.config.DemoUsername == "" {
		http.Error(w, "password login disabled", http.StatusNotFound)
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decode(w, r, &input) {
		return
	}
	// Perform the same password verification even for an unknown username.
	valid, err := h.hasher.Verify(r.Context(), []byte(input.Password), h.config.DemoPasswordHash)
	if err != nil {
		failure(w, err)
		return
	}
	if !valid || subtle.ConstantTimeCompare([]byte(input.Username), []byte(h.config.DemoUsername)) != 1 {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	p := identity.Principal{ID: "password:" + h.config.DemoUsername, Name: h.config.DemoUsername}
	if err := h.login.SignIn(r.Context(), p); err != nil {
		failure(w, err)
		return
	}
	respond(w, http.StatusOK, p)
}

func (h *Workspace) list(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.FromContext[identity.Principal](r.Context())
	items, err := h.queries.ListNotes.Handle(r.Context(), notequery.ListNotesQuery{Owner: p.ID})
	if err != nil {
		failure(w, err)
		return
	}
	respond(w, http.StatusOK, items)
}
func (h *Workspace) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &input) {
		return
	}
	p, _ := auth.FromContext[identity.Principal](r.Context())
	note, err := h.commands.CreateNote.Handle(r.Context(), notecommand.CreateNoteCommand{Owner: p.ID, Text: input.Text})
	if err != nil {
		failure(w, err)
		return
	}
	respond(w, http.StatusCreated, note)
}
func (h *Workspace) update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &input) {
		return
	}
	p, _ := auth.FromContext[identity.Principal](r.Context())
	if _, err := h.commands.UpdateNote.Handle(r.Context(), notecommand.UpdateNoteCommand{Owner: p.ID, ID: r.PathValue("id"), Text: input.Text}); err != nil {
		failure(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Workspace) delete(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.FromContext[identity.Principal](r.Context())
	if _, err := h.commands.DeleteNote.Handle(r.Context(), notecommand.DeleteNoteCommand{Owner: p.ID, ID: r.PathValue("id")}); err != nil {
		failure(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Workspace) registerObjects(mux *http.ServeMux, store *s3x.Store) {
	mux.HandleFunc("POST /v1/objects", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
		if err != nil {
			http.Error(w, "object exceeds 64 KiB", http.StatusRequestEntityTooLarge)
			return
		}
		key := objectPrefix(r) + uuid.NewV7().String()
		if _, err := store.Put(r.Context(), key, strings.NewReader(string(body)), int64(len(body)), "application/octet-stream"); err != nil {
			failure(w, err)
			return
		}
		respond(w, http.StatusCreated, map[string]string{"key": key})
	})
	mux.HandleFunc("GET /v1/objects", func(w http.ResponseWriter, r *http.Request) {
		page, err := store.List(r.Context(), objectPrefix(r), r.URL.Query().Get("cursor"), 100)
		if err != nil {
			failure(w, err)
			return
		}
		keys := make([]string, 0, len(page.Contents))
		for _, object := range page.Contents {
			keys = append(keys, aws.ToString(object.Key))
		}
		respond(w, http.StatusOK, struct {
			Keys   []string `json:"keys"`
			Cursor string   `json:"cursor"`
		}{keys, aws.ToString(page.NextContinuationToken)})
	})
	mux.HandleFunc("GET /v1/objects/content", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !strings.HasPrefix(key, objectPrefix(r)) {
			http.Error(w, "object not found", http.StatusNotFound)
			return
		}
		object, err := store.Get(r.Context(), key)
		if err != nil {
			failure(w, err)
			return
		}
		defer object.Body.Close()
		if object.ContentLength != nil && *object.ContentLength > 64<<10 {
			http.Error(w, "object exceeds 64 KiB", http.StatusRequestEntityTooLarge)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = io.Copy(w, io.LimitReader(object.Body, 64<<10))
	})
	mux.HandleFunc("DELETE /v1/objects/content", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !strings.HasPrefix(key, objectPrefix(r)) {
			http.Error(w, "object not found", http.StatusNotFound)
			return
		}
		if err := store.Delete(r.Context(), key); err != nil {
			failure(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
func objectPrefix(r *http.Request) string {
	p, _ := auth.FromContext[identity.Principal](r.Context())
	digest := sha256.Sum256([]byte(p.ID))
	return hex.EncodeToString(digest[:]) + "/"
}
func decode(w http.ResponseWriter, r *http.Request, destination any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		http.Error(w, "one JSON value required", http.StatusBadRequest)
		return false
	}
	return true
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func failure(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, notes.ErrInvalid) {
		status = http.StatusBadRequest
	}
	if errors.Is(err, notes.ErrNotFound) {
		status = http.StatusNotFound
	}
	http.Error(w, http.StatusText(status), status)
}

// Refresh records an idempotent, bounded cache update rather than a business write.
func (h *Workspace) Refresh(ctx context.Context, _ any) (any, error) {
	return nil, h.cache.Set(ctx, "refresh:last-run", []byte(time.Now().UTC().Format(time.RFC3339Nano)), 24*time.Hour)
}
