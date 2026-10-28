// Package user is the composition root of the user service.
package user

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/health"
	"github.com/nostosx/servex/transport/httpx"

	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/server"
	httpport "github.com/nostosx/servex-examples/services/user/internal/port/http"

	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	rpcport "github.com/nostosx/servex-examples/services/user/internal/port/rpc"
)

// Run assembles this service and owns its transport lifecycle.
func Run(ctx context.Context, cfg config.Config) error {
	var application *app.App
	check := health.CheckerFunc(func(ctx context.Context) error { return application.Check(ctx) })
	servers, err := server.New(ctx, cfg, check)
	if err != nil {
		return err
	}
	Register(servers.HTTP, servers.GRPC)
	application = app.New(app.WithHooks(servers.HTTP, servers.GRPC), app.WithStopTimeout(10*time.Second))
	return application.Run(ctx)
}

// Register mounts transport handlers. Construct fallible dependencies before calling it.
func Register(http *httpx.Server, registrar grpc.ServiceRegistrar) {
	httpport.Register(http)
	pb.RegisterUserServiceServer(registrar, &rpcport.Handler{})
}
