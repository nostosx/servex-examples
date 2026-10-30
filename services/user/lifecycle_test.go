package user

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nostosx/servex-contrib/lifecycle"
	"github.com/nostosx/servex/app"
)

type blockedConsumer struct {
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (*blockedConsumer) Name() string                { return "consumer" }
func (*blockedConsumer) Start(context.Context) error { return nil }
func (h *blockedConsumer) Stop(context.Context) error {
	h.calls.Add(1)
	close(h.started)
	<-h.release
	return nil
}

func TestRuntimeCloseWaitsForConsumerAndNativeRequest(t *testing.T) {
	scope, err := lifecycle.NewScope(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	closed := make(chan struct{})
	resource, err := lifecycle.New("shared-resource", 0, lifecycle.Operations[int]{Close: func(context.Context, int) error { close(closed); return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err := scope.Add(resource); err != nil {
		t.Fatal(err)
	}
	gate := newRequestGate()
	consumer := &blockedConsumer{started: make(chan struct{}), release: make(chan struct{})}
	guard, hooks := guardResources(scope, gate, []app.Hook{consumer})
	runtime := &Runtime{App: app.New(app.WithHooks(append([]app.Hook{guard}, hooks...)...)), guard: guard}
	if err := runtime.App.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	entered, releaseRequest := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(gate.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-releaseRequest
		w.WriteHeader(http.StatusNoContent)
	})))
	t.Cleanup(func() {
		select {
		case <-consumer.release:
		default:
			close(consumer.release)
		}
		select {
		case <-releaseRequest:
		default:
			close(releaseRequest)
		}
		server.Close()
		closeRuntime(t, runtime)
	})
	responseDone := make(chan error, 1)
	go func() {
		response, err := server.Client().Get(server.URL)
		if err == nil {
			_, err = io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
		responseDone <- err
	}()
	<-entered
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	err = runtime.Close(ctx)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("short close: %v", err)
	}
	<-consumer.started
	select {
	case <-closed:
		t.Fatal("resource closed while consumer running")
	default:
	}
	close(consumer.release)
	// The consumer has stopped, but the native handler still holds its dependency.
	ctx, cancel = context.WithTimeout(t.Context(), 10*time.Millisecond)
	err = runtime.Close(ctx)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("handler close: %v", err)
	}
	select {
	case <-closed:
		t.Fatal("resource closed while native handler running")
	default:
	}
	close(releaseRequest)
	if err := <-responseDone; err != nil {
		t.Fatal(err)
	}
	closeRuntime(t, runtime)
	select {
	case <-closed:
	default:
		t.Fatal("resource not closed after consumers joined")
	}
	if consumer.calls.Load() != 1 {
		t.Fatalf("native Stop called %d times", consumer.calls.Load())
	}
}
