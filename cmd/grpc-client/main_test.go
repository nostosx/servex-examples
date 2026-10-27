package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/nostosx/servex-contrib/transport/grpcx"
	"github.com/nostosx/servex/errorx"
	"github.com/nostosx/servex/health"
)

func TestCheckHealth(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		ready      bool
		wantStatus grpc_health_v1.HealthCheckResponse_ServingStatus
	}{
		{name: "serving", ready: true, wantStatus: grpc_health_v1.HealthCheckResponse_SERVING},
		{name: "not serving", wantStatus: grpc_health_v1.HealthCheckResponse_NOT_SERVING},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			srv := grpcx.NewServer(grpcx.WithAddress("127.0.0.1:0"), grpcx.WithHealth(health.CheckerFunc(func(context.Context) error {
				if !tt.ready {
					return errors.New("not ready")
				}
				return nil
			})))
			if err := srv.Start(ctx); err != nil {
				t.Fatal(err)
			}
			defer func() {
				stop, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if err := srv.Stop(stop); err != nil {
					t.Error(err)
				}
			}()
			status, err := checkHealth(ctx, srv.URL().String())
			if status != tt.wantStatus {
				t.Errorf("status = %s, want %s", status, tt.wantStatus)
			}
			if (err == nil) != tt.ready {
				t.Fatalf("ready = %t, error = %v", tt.ready, err)
			}
		})
	}
}

func TestCheckHealthCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := checkHealth(ctx, "grpc://127.0.0.1:1"); !errors.Is(err, context.Canceled) || errorx.CodeOf(err) != errorx.Canceled {
		t.Fatalf("error = %v, want cancellation", err)
	}
}
