package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"errors"

	"github.com/nostosx/servex-contrib/transport/httpx"
	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	"github.com/nostosx/servex-examples/application/notes"
	notecommand "github.com/nostosx/servex-examples/application/notes/command"
	notequery "github.com/nostosx/servex-examples/application/notes/query"
	"github.com/nostosx/servex-examples/application/user/query"
	"github.com/nostosx/servex/cqrs"
)

func TestUserTransportsShareCQRSChain(t *testing.T) {
	t.Parallel()
	type contextKey struct{}
	calls := 0
	assembled := 0
	middleware := func(meta cqrs.Metadata, next cqrs.Handler[query.GetUserQuery, query.GetUserResult]) cqrs.Handler[query.GetUserQuery, query.GetUserResult] {
		assembled++
		if meta.Kind != cqrs.KindQuery || meta.Operation != "user.get_user" {
			t.Fatalf("metadata=%+v", meta)
		}
		return cqrs.QueryFunc[query.GetUserQuery, query.GetUserResult](func(ctx context.Context, in query.GetUserQuery) (query.GetUserResult, error) {
			calls++
			if ctx.Value(contextKey{}) != "request" {
				t.Fatal("context lost")
			}
			return next.Handle(ctx, in)
		})
	}
	server := httpx.NewServer()
	var registrar registry
	service := notes.NewService(notes.NewMemoryRepository())
	Register(server, &registrar, newCommand(nil)(service), newQuery(nil, middleware)(service))
	ctx := context.WithValue(t.Context(), contextKey{}, "request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/users/7", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("HTTP %d: %s", response.Code, response.Body)
	}
	rpc := registrar.implementation.(pb.UserServiceServer)
	reply, err := rpc.GetUser(ctx, &pb.GetUserRequest{Id: 7})
	if err != nil || reply.GetName() != "alice" {
		t.Fatalf("RPC %v %v", reply, err)
	}
	if assembled != 1 || calls != 2 {
		t.Fatalf("assembled=%d calls=%d", assembled, calls)
	}
}

func TestNoteCommandsPreserveOwnerIsolation(t *testing.T) {
	t.Parallel()
	// Exercise the actual handlers supplied by service assembly.
	repository := notes.NewService(notes.NewMemoryRepository())
	commands := newCommand(nil)(repository)
	queries := newQuery(nil)(repository)
	created, err := commands.CreateNote.Handle(t.Context(), notecommand.CreateNoteCommand{Owner: "owner-a", Text: "private"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = commands.DeleteNote.Handle(t.Context(), notecommand.DeleteNoteCommand{Owner: "owner-b", ID: created.ID})
	if !errors.Is(err, notes.ErrNotFound) {
		t.Fatal("cross-owner delete succeeded")
	}
	items, err := queries.ListNotes.Handle(t.Context(), notequery.ListNotesQuery{Owner: "owner-a"})
	if err != nil || len(items) != 1 || !reflect.DeepEqual(items[0], created) {
		t.Fatalf("items=%v err=%v", items, err)
	}
}
