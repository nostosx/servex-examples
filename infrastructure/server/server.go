// Package server constructs the process transports without starting them.
package server

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/nostosx/servex-contrib/transport/grpcx"
	"github.com/nostosx/servex-contrib/transport/httpx"
	"github.com/nostosx/servex/health"
	"github.com/nostosx/servex/logger"

	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/rpc"
)

// Servers contains only the listeners enabled for this project.
type Servers struct {
	HTTP *httpx.Server
	GRPC *grpcx.Server
}

// New validates configuration and constructs transports. The caller mounts
// handlers and passes the transports to app.WithHooks in explicit order.
func New(ctx context.Context, cfg config.Config, check health.Checker, log logger.Logger) (*Servers, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	http := httpx.NewServer(httpx.WithLogger(log), httpx.WithAddress(cfg.Address), httpx.WithHealth(check),
		httpx.WithReadHeaderTimeout(5*time.Second), httpx.WithReadTimeout(10*time.Second),
		httpx.WithWriteTimeout(10*time.Second), httpx.WithIdleTimeout(time.Minute), httpx.WithMaxHeaderBytes(16<<10), httpx.WithMaxBodyBytes(64<<10))
	rpcServer := grpcx.NewServer(grpcx.WithAddress(cfg.GRPCAddress), grpcx.WithHealth(check),
		grpcx.WithGRPCOptions(grpc.ChainUnaryInterceptor(rpc.Authenticate(cfg.Token, log)), grpc.ChainStreamInterceptor(rpc.AuthenticateStream(cfg.Token, log)), grpc.MaxRecvMsgSize(rpc.MaxMessageBytes), grpc.MaxSendMsgSize(rpc.MaxMessageBytes)))
	return &Servers{HTTP: http, GRPC: rpcServer}, nil
}
