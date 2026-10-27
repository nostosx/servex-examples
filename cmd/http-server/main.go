// Command http-server demonstrates the servex HTTP server.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/errorx"
	"github.com/nostosx/servex/health"
	"github.com/nostosx/servex/transport/httpx"
)

type getUserRequest struct {
	ID int `json:"id"`
}

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getUser(_ context.Context, r *getUserRequest) (*user, error) {
	if r.ID <= 0 {
		return nil, errorx.New(errorx.InvalidArgument, "INVALID_ID", "id must be positive")
	}
	return &user{ID: r.ID, Name: "alice"}, nil
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:8000", "listen address")
	flag.Parse()
	a, _ := newApp(*addr)
	ctx, stop := app.SignalContext(context.Background())
	defer stop()
	return a.Run(ctx)
}

func newApp(addr string) (*app.App, *httpx.Server) {
	var a *app.App
	srv := httpx.NewServer(
		httpx.WithAddress(addr),
		httpx.WithReadHeaderTimeout(5*time.Second),
		httpx.WithReadTimeout(10*time.Second),
		httpx.WithWriteTimeout(10*time.Second),
		httpx.WithIdleTimeout(time.Minute),
		httpx.WithMaxBodyBytes(64<<10),
		// Readiness follows the application lifecycle through a.Check.
		httpx.WithHealth(health.CheckerFunc(func(ctx context.Context) error { return a.Check(ctx) })),
	)
	// Route parameters bind by JSON field name and take precedence over query values.
	// RouteFunc keeps the request type and the handler in step.
	httpx.RouteFunc(srv, "GET /v1/users/{id}", getUser)

	a = app.New(app.WithHooks(srv))
	return a, srv
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
