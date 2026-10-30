package user

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/nostosx/servex-contrib/lifecycle"
	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/health"
)

// ownedHook makes repeated Stop calls wait for the actual native shutdown.
// Caller deadlines bound waiting; they never allow dependencies to close early.
type ownedHook struct {
	app.Hook
	once sync.Once
	done chan struct{}
	err  error
}

func (h *ownedHook) Stop(ctx context.Context) error {
	h.once.Do(func() {
		go func() {
			stop, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			h.err = h.Hook.Stop(stop)
			cancel()
			// Worker and scheduler hooks can return on their deadline while still
			// draining. Repeated Stop joins them before closing shared connections.
			if errors.Is(h.err, context.DeadlineExceeded) || errors.Is(h.err, context.Canceled) {
				h.err = errors.Join(h.err, h.Hook.Stop(context.Background()))
			}
			close(h.done)
		}()
	})
	select {
	case <-h.done:
		return h.err
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
func (h *ownedHook) Check(ctx context.Context) error {
	if checker, ok := h.Hook.(health.Checker); ok {
		return checker.Check(ctx)
	}
	return context.Cause(ctx)
}
func (h *ownedHook) Err() <-chan error {
	if reporter, ok := h.Hook.(interface{ Err() <-chan error }); ok {
		return reporter.Err()
	}
	return nil
}

type resourceGuard struct {
	scope     *lifecycle.Scope
	requests  *requestGate
	consumers []*ownedHook
	once      sync.Once
	done      chan struct{}
	err       error
}

func guardResources(scope *lifecycle.Scope, requests *requestGate, hooks []app.Hook) (*resourceGuard, []app.Hook) {
	guard := &resourceGuard{scope: scope, requests: requests, done: make(chan struct{})}
	owned := make([]app.Hook, 0, len(hooks))
	for _, hook := range hooks {
		consumer := &ownedHook{Hook: hook, done: make(chan struct{})}
		guard.consumers = append(guard.consumers, consumer)
		owned = append(owned, consumer)
	}
	return guard, owned
}
func (g *resourceGuard) Name() string                    { return g.scope.Name() }
func (g *resourceGuard) Start(ctx context.Context) error { return g.scope.Start(ctx) }
func (g *resourceGuard) Check(ctx context.Context) error { return g.scope.Check(ctx) }
func (g *resourceGuard) Stop(ctx context.Context) error {
	g.once.Do(func() {
		g.requests.close()
		go func() {
			for _, hook := range slices.Backward(g.consumers) {
				g.err = errors.Join(g.err, hook.Stop(context.Background()))
			}
			g.requests.wait()
			g.err = errors.Join(g.err, g.scope.Stop(context.Background()))
			close(g.done)
		}()
	})
	select {
	case <-g.done:
		return g.err
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

// requestGate joins native handlers even if the transport force-closes a socket.
type requestGate struct {
	mu      sync.Mutex
	closing bool
	active  sync.WaitGroup
}

func newRequestGate() *requestGate { return &requestGate{} }
func (g *requestGate) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.mu.Lock()
		if g.closing {
			g.mu.Unlock()
			http.Error(w, "server closing", http.StatusServiceUnavailable)
			return
		}
		g.active.Add(1)
		g.mu.Unlock()
		defer g.active.Done()
		next.ServeHTTP(w, r)
	})
}
func (g *requestGate) close() { g.mu.Lock(); g.closing = true; g.mu.Unlock() }
func (g *requestGate) wait()  { g.active.Wait() }
