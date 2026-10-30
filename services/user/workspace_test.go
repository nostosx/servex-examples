package user

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nostosx/servex-contrib/auth/password"
	"github.com/nostosx/servex-examples/application/notes"
	"github.com/nostosx/servex-examples/infrastructure/config"
)

func workspaceConfig(t *testing.T) config.Config {
	t.Helper()
	hasher, err := password.New(password.Parameters{Memory: 8, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 16})
	if err != nil {
		t.Fatal(err)
	}
	hash, err := hasher.Hash(t.Context(), []byte("test-only-password"))
	if err != nil {
		t.Fatal(err)
	}
	return config.Config{Address: "127.0.0.1:0", GRPCAddress: "127.0.0.1:0", Token: strings.Repeat("a", 32), LogLevel: "error", Composition: config.Composition{DemoUsername: "demo", DemoPasswordHash: hash, InsecureCookies: true, Namespace: "test-example:"}}
}
func closeRuntime(t *testing.T, runtime *Runtime) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := runtime.Close(ctx); err != nil {
		t.Error(err)
	}
}
func workspaceClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar, Timeout: 10 * time.Second}
}
func request(t *testing.T, client *http.Client, method, target, body string, status int) []byte {
	t.Helper()
	return requestContext(t, t.Context(), client, method, target, body, status)
}
func requestContext(t *testing.T, ctx context.Context, client *http.Client, method, target, body string, status int) []byte {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, method, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Fatalf("%s %s: status=%d want=%d body=%s", method, target, response.StatusCode, status, data)
	}
	return data
}
func signIn(t *testing.T, client *http.Client, base string) {
	t.Helper()
	request(t, client, http.MethodPost, base+"/auth/password", `{"username":"demo","password":"test-only-password"}`, http.StatusOK)
}

func TestWorkspaceLoginCRUDAndRevocation(t *testing.T) {
	cfg := workspaceConfig(t)
	runtime, err := Build(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeRuntime(t, runtime) })
	server := httptest.NewServer(runtime.Servers.HTTP)
	t.Cleanup(server.Close)
	client := workspaceClient(t)
	request(t, client, http.MethodGet, server.URL+"/v1/notes", "", http.StatusUnauthorized)
	request(t, client, http.MethodGet, server.URL+"/v1/users/7", "", http.StatusOK)
	request(t, client, http.MethodPost, server.URL+"/auth/password", `{"username":"demo","password":"wrong"}`, http.StatusUnauthorized)
	signIn(t, client, server.URL)
	old := client.Jar.Cookies(mustURL(t, server.URL))
	if len(old) != 1 {
		t.Fatalf("expected one opaque session cookie, got %v", old)
	}
	created := request(t, client, http.MethodPost, server.URL+"/v1/notes", `{"text":"persist across requests"}`, http.StatusCreated)
	var note notes.Note
	if err := json.Unmarshal(created, &note); err != nil {
		t.Fatal(err)
	}
	if note.ID == "" || note.CreatedAt.IsZero() {
		t.Fatal(note)
	}
	request(t, client, http.MethodPut, server.URL+"/v1/notes/"+note.ID, `{"text":"updated"}`, http.StatusNoContent)
	listed := request(t, client, http.MethodGet, server.URL+"/v1/notes", "", http.StatusOK)
	if !strings.Contains(string(listed), `"text":"updated"`) {
		t.Fatal(string(listed))
	}
	request(t, client, http.MethodDelete, server.URL+"/v1/notes/"+note.ID, "", http.StatusNoContent)
	request(t, client, http.MethodDelete, server.URL+"/v1/notes/"+note.ID, "", http.StatusNotFound)
	request(t, client, http.MethodPost, server.URL+"/auth/logout", "", http.StatusNoContent)
	request(t, client, http.MethodGet, server.URL+"/v1/notes", "", http.StatusUnauthorized)
	client.Jar.SetCookies(mustURL(t, server.URL), old)
	request(t, client, http.MethodGet, server.URL+"/v1/me", "", http.StatusUnauthorized)
	metrics := request(t, client, http.MethodGet, server.URL+"/metrics", "", http.StatusOK)
	if !strings.Contains(string(metrics), `operation="http.request"`) {
		t.Fatalf("native HTTP metrics missing: %s", metrics)
	}
}

func TestWorkspaceCSRFBeforeSessionAndLoginRotation(t *testing.T) {
	runtime, err := Build(t.Context(), workspaceConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeRuntime(t, runtime) })
	server := httptest.NewServer(runtime.Servers.HTTP)
	t.Cleanup(server.Close)
	client := workspaceClient(t)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, server.URL+"/auth/password", strings.NewReader(`{"username":"demo","password":"test-only-password"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "https://untrusted.example")
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden || len(response.Cookies()) != 0 {
		t.Fatalf("cross-origin request reached session: %d %v", response.StatusCode, response.Cookies())
	}
	signIn(t, client, server.URL)
	old := client.Jar.Cookies(mustURL(t, server.URL))
	signIn(t, client, server.URL)
	rotated := client.Jar.Cookies(mustURL(t, server.URL))
	if len(old) != 1 || len(rotated) != 1 || old[0].Value == rotated[0].Value {
		t.Fatal("sign-in did not rotate opaque token")
	}
	replay := workspaceClient(t)
	replay.Jar.SetCookies(mustURL(t, server.URL), old)
	request(t, replay, http.MethodGet, server.URL+"/v1/me", "", http.StatusUnauthorized)
	request(t, client, http.MethodGet, server.URL+"/v1/me", "", http.StatusOK)
}

func TestRuntimeCloseBeforeStartAndFailedAssembly(t *testing.T) {
	runtime, err := Build(t.Context(), workspaceConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	closeRuntime(t, runtime)
	closeRuntime(t, runtime)
	if err := runtime.App.Start(t.Context()); err == nil {
		t.Fatal("closed runtime restarted")
	}
	cfg := workspaceConfig(t)
	cfg.Composition.DemoPasswordHash = "malformed"
	if runtime, err := Build(t.Context(), cfg); err == nil || runtime != nil {
		t.Fatalf("bad password hash accepted: %v %v", runtime, err)
	}
}

func mustURL(t *testing.T, value string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
