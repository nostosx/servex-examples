package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	serviceapp "github.com/nostosx/servex-examples/services/user/internal/application"

	"github.com/nostosx/servex-examples/application/user/query"

	"github.com/nostosx/servex-contrib/transport/httpx"
)

func TestRoutes(t *testing.T) {
	t.Parallel()
	srv := httpx.NewServer(httpx.WithMaxBodyBytes(64 << 10))
	Register(srv, &UserServer{queries: &serviceapp.Query{GetUser: query.NewGetUserHandler()}})
	for _, tt := range []struct {
		name, method, path, body string
		unknownLength            bool
		wantStatus               int
		wantReason               string
	}{
		{name: "path ID overrides query", method: http.MethodGet, path: "/v1/users/7?id=9", wantStatus: http.StatusOK},
		{name: "zero ID", method: http.MethodGet, path: "/v1/users/0", wantStatus: http.StatusBadRequest, wantReason: "INVALID_ID"},
		{name: "negative ID", method: http.MethodGet, path: "/v1/users/-1", wantStatus: http.StatusBadRequest, wantReason: "INVALID_ID"},
		{name: "malformed ID", method: http.MethodGet, path: "/v1/users/nope", wantStatus: http.StatusBadRequest, wantReason: "INVALID_PARAMETER"},
		{name: "unsupported method", method: http.MethodPost, path: "/v1/users/7", wantStatus: http.StatusMethodNotAllowed},
		{name: "oversized body", method: http.MethodGet, path: "/v1/users/7", body: strings.Repeat(" ", (64<<10)+1), wantStatus: http.StatusRequestEntityTooLarge, wantReason: "BODY_TOO_LARGE"},
		{name: "unknown body length", method: http.MethodGet, path: "/v1/users/7", body: strings.Repeat(" ", (64<<10)+1), unknownLength: true, wantStatus: http.StatusRequestEntityTooLarge, wantReason: "BODY_TOO_LARGE"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.unknownLength {
				req.ContentLength = -1
			}
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tt.wantStatus, w.Body)
			}
			if tt.wantStatus == http.StatusOK {
				var got GetUserResponse
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.ID != 7 || got.Name != "alice" {
					t.Fatalf("user=%+v", got)
				}
			} else if tt.wantReason != "" {
				var got struct{ Reason string }
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Reason != tt.wantReason {
					t.Fatalf("reason = %q, want %q", got.Reason, tt.wantReason)
				}
			}
		})
	}
}
