package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nostosx/servex/errorx"
)

func TestFetchUser(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body, wantReason string
		responseStatus         int
		wantError              bool
	}{
		{name: "success", body: `{"id":7,"name":"alice"}`, responseStatus: http.StatusOK},
		{name: "structured failure", body: `{"code":"InvalidArgument","reason":"INVALID_ID","message":"bad ID"}`, responseStatus: http.StatusBadRequest, wantError: true, wantReason: "INVALID_ID"},
		{name: "malformed response", body: "{", responseStatus: http.StatusOK, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/users/7" {
					t.Errorf("request=%s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			got, err := fetchUser(ctx, srv.URL)
			if tt.wantError {
				if err == nil {
					t.Fatal("invalid response succeeded")
				}
				if tt.wantReason != "" {
					gotError := errorx.From(err)
					if gotError.Code != errorx.InvalidArgument || gotError.Reason != tt.wantReason {
						t.Fatalf("error = %v, want InvalidArgument with reason %s", err, tt.wantReason)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got == nil || got.ID != 7 || got.Name != "alice" {
				t.Fatalf("user = %+v, want {ID:7 Name:alice}", got)
			}
		})
	}
}

func TestFetchUserCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := fetchUser(ctx, "http://127.0.0.1:1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
