// Command grpc-server exposes application readiness through the gRPC health API.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/nostosx/servex-contrib/transport/grpcx"
	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/health"
)

func run() error {
	addr := flag.String("addr", "127.0.0.1:9000", "listen address")
	flag.Parse()
	a, _ := newApp(*addr)
	ctx, stop := app.SignalContext(context.Background())
	defer stop()
	return a.Run(ctx)
}

func newApp(addr string) (*app.App, *grpcx.Server) {
	var a *app.App
	srv := grpcx.NewServer(
		grpcx.WithAddress(addr),
		grpcx.WithHealth(health.CheckerFunc(func(ctx context.Context) error { return a.Check(ctx) })),
		// grpcx.WithMiddleware(...) adds servex server middleware for unary calls.
	)

	a = app.New(app.WithHooks(srv))
	return a, srv
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
