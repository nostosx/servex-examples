// Package user is the composition root of the user service.
package user

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corssdk "github.com/rs/cors"
	"google.golang.org/grpc"

	"github.com/nostosx/servex-contrib/auth/login"
	"github.com/nostosx/servex-contrib/auth/password"
	dox "github.com/nostosx/servex-contrib/di/do"
	"github.com/nostosx/servex-contrib/lifecycle"
	prometheusx "github.com/nostosx/servex-contrib/metrics/prometheus"
	"github.com/nostosx/servex-contrib/middleware/profile"
	corsx "github.com/nostosx/servex-contrib/transport/cors"
	"github.com/nostosx/servex-contrib/transport/httpx"
	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/di"
	"github.com/nostosx/servex/endpoint"
	"github.com/nostosx/servex/health"
	"github.com/nostosx/servex/messaging/jobqueue"
	"github.com/nostosx/servex/storage/cache"

	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	"github.com/nostosx/servex-examples/application/identity"
	"github.com/nostosx/servex-examples/application/notes"
	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/logging"
	"github.com/nostosx/servex-examples/infrastructure/server"
	httpport "github.com/nostosx/servex-examples/services/user/internal/port/http"
	rpcport "github.com/nostosx/servex-examples/services/user/internal/port/rpc"
)

// Runtime owns resources even when callers never start App. Close after use.
type Runtime struct {
	App       *app.App
	Servers   *server.Servers
	guard     *resourceGuard
	closeOnce sync.Once
	closeDone chan struct{}
	closeErr  error
}

// Close stops consumers before shared resources. It also works before App.Start.
func (r *Runtime) Close(ctx context.Context) error {
	r.guard.requests.close()
	r.closeOnce.Do(func() {
		r.closeDone = make(chan struct{})
		go func() {
			r.closeErr = errors.Join(r.App.Stop(context.Background()), r.guard.Stop(context.Background()))
			close(r.closeDone)
		}()
	})
	select {
	case <-r.closeDone:
		return r.closeErr
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

// Build assembles dependencies without starting listeners or migrating schemas.
func Build(ctx context.Context, cfg config.Config) (_ *Runtime, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	scope, err := lifecycle.NewScope(10 * time.Second)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			scope.Finish(&err)
		}
	}()
	log, closeLog, err := logging.New(cfg.LogLevel, nil, cfg.LogFile)
	if err != nil {
		return nil, err
	}
	logResource, err := lifecycle.New("logger", closeLog, lifecycle.Operations[func() error]{Close: func(_ context.Context, close func() error) error { return close() }})
	if err != nil {
		return nil, err
	}
	if err := scope.Add(logResource); err != nil {
		return nil, err
	}
	registry := prom.NewRegistry()
	meter := prometheusx.New(registry, prometheusx.WithLogger(log))
	signals := profile.New(profile.Config{Logger: log, Meter: meter})
	resources, err := buildResources(ctx, cfg.Composition, scope, meter)
	if err != nil {
		return nil, err
	}
	manager, err := buildLogin(ctx, cfg.Composition, resources.sessions, resources.states)
	if err != nil {
		return nil, err
	}
	hasher, err := password.New(password.DefaultParameters())
	if err != nil {
		return nil, err
	}
	if cfg.Composition.DemoPasswordHash != "" {
		if _, err := hasher.NeedsRehash(cfg.Composition.DemoPasswordHash); err != nil {
			return nil, err
		}
	}
	// The scope owns SDK resources. The container owns only the constructor graph.
	container := dox.New()
	if err := scope.Add(container); err != nil {
		return nil, err
	}
	if err := di.Supply[notes.Repository](container, resources.repository); err != nil {
		return nil, err
	}
	if err := di.Supply[*login.Manager[identity.Principal]](container, manager); err != nil {
		return nil, err
	}
	if err := di.Supply[*password.Hasher](container, hasher); err != nil {
		return nil, err
	}
	if err := di.Supply[config.Composition](container, cfg.Composition); err != nil {
		return nil, err
	}
	if err := di.Supply[cache.Cache](container, resources.cache); err != nil {
		return nil, err
	}
	if err := dox.ProvideConstructors(container, notes.NewService, httpport.NewWorkspace); err != nil {
		return nil, err
	}
	workspace, err := di.Resolve[*httpport.Workspace](container)
	if err != nil {
		return nil, err
	}
	requests := newRequestGate()
	outer := []func(http.Handler) http.Handler{requests.middleware, signals.HTTP("http.request")}
	if len(cfg.Composition.AllowedOrigins) > 0 {
		cors, err := corsx.New(corssdk.Options{AllowedOrigins: cfg.Composition.AllowedOrigins, AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"}, AllowedHeaders: []string{"Content-Type"}, AllowCredentials: true})
		if err != nil {
			return nil, err
		}
		outer = append(outer, cors)
	}
	csrf := http.NewCrossOriginProtection()
	// TrustedOrigins authorizes cross-origin writes independently of CORS.
	for _, origin := range cfg.Composition.AllowedOrigins {
		if err := csrf.AddTrustedOrigin(origin); err != nil {
			return nil, err
		}
	}
	outer = append(outer, csrf.Handler, manager.Middleware)
	runtime := &Runtime{}
	check := health.CheckerFunc(func(ctx context.Context) error { return runtime.App.Check(ctx) })
	servers, err := server.New(ctx, cfg, check, log, server.Options{HTTP: outer, Unary: signals.Server(), Stream: signals.Stream()})
	if err != nil {
		return nil, err
	}
	runtime.Servers = servers
	Register(servers.HTTP, servers.GRPC)
	var enqueuer jobqueue.Enqueuer
	if resources.queue != nil {
		enqueuer = resources.queue
	}
	workspace.Register(servers.HTTP, enqueuer, resources.objects, resources.coordinator)
	servers.HTTP.HandleHTTP("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	hooks := []app.Hook{}
	if resources.queue != nil {
		hooks = append(hooks, jobqueue.NewWorker(resources.queue, map[string]endpoint.Endpoint{"example.refresh": workspace.Refresh}, jobqueue.WithMiddleware(signals.Server()...)))
	}
	if resources.schedule != nil {
		hooks = append(hooks, resources.schedule)
	}
	hooks = append(hooks, servers.HTTP, servers.GRPC)
	guard, owned := guardResources(scope, requests, hooks)
	runtime.guard = guard
	runtime.App = app.New(app.WithLogger(log), app.WithHooks(append([]app.Hook{guard}, owned...)...), app.WithStopTimeout(10*time.Second))
	return runtime, nil
}

// Run owns assembly, serving and cleanup with an uncanceled shutdown context.
func Run(ctx context.Context, cfg config.Config) (err error) {
	runtime, err := Build(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = errors.Join(err, runtime.Close(cleanup))
	}()
	return runtime.App.Run(ctx)
}

// Register mounts the original public user route and token-protected gRPC API.
func Register(http *httpx.Server, registrar grpc.ServiceRegistrar) {
	httpport.Register(http)
	pb.RegisterUserServiceServer(registrar, &rpcport.Handler{})
}
