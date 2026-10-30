// Package server constructs the process transports without starting them.
package server

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"github.com/nostosx/servex-contrib/transport/grpcx"
	"github.com/nostosx/servex-contrib/transport/httpx"
	"github.com/nostosx/servex/endpoint"
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

// Options keeps native HTTP scope separate from unary and stream middleware.
type Options struct {
	HTTP   []func(http.Handler) http.Handler
	Unary  []endpoint.Middleware
	Stream []endpoint.Middleware
}

// New validates configuration and constructs transports. The caller mounts
// handlers and passes the transports to app.WithHooks in explicit order.
func New(ctx context.Context, cfg config.Config, check health.Checker, log logger.Logger, options ...Options) (*Servers, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	var o Options
	for _, option := range options {
		o.HTTP = append(o.HTTP, option.HTTP...)
		o.Unary = append(o.Unary, option.Unary...)
		o.Stream = append(o.Stream, option.Stream...)
	}
	http := httpx.NewServer(httpx.WithHTTPMiddleware(o.HTTP...), httpx.WithLogger(log), httpx.WithAddress(cfg.Address), httpx.WithHealth(check),
		httpx.WithReadHeaderTimeout(5*time.Second), httpx.WithReadTimeout(10*time.Second),
		httpx.WithWriteTimeout(10*time.Second), httpx.WithIdleTimeout(time.Minute), httpx.WithMaxHeaderBytes(16<<10), httpx.WithMaxBodyBytes(64<<10))
	rpcServer := grpcx.NewServer(grpcx.WithMiddleware(o.Unary...), grpcx.WithStreamMiddleware(o.Stream...), grpcx.WithAddress(cfg.GRPCAddress), grpcx.WithHealth(check),
		grpcx.WithGRPCOptions(grpc.ChainUnaryInterceptor(rpc.Authenticate(cfg.Token, log)), grpc.ChainStreamInterceptor(rpc.AuthenticateStream(cfg.Token, log)), grpc.MaxRecvMsgSize(rpc.MaxMessageBytes), grpc.MaxSendMsgSize(rpc.MaxMessageBytes)))
	return &Servers{HTTP: http, GRPC: rpcServer}, nil
}
