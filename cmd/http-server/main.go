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

func getUser(_ context.Context, req any) (any, error) {
	r := req.(*getUserRequest)
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
		httpx.Address(addr),
		httpx.ReadHeaderTimeout(5*time.Second),
		httpx.ReadTimeout(10*time.Second),
		httpx.WriteTimeout(10*time.Second),
		httpx.IdleTimeout(time.Minute),
		httpx.MaxBodyBytes(64<<10),
		// Readiness follows the application lifecycle through a.Check.
		httpx.Health(health.CheckerFunc(func(ctx context.Context) error { return a.Check(ctx) })),
	)
	// Route parameters bind by JSON field name and take precedence over query values.
	httpx.Route[getUserRequest](srv, "GET /v1/users/{id}", getUser)

	a = app.New(app.WithHooks(srv))
	return a, srv
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
