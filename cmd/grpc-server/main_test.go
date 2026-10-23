package main

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/nostosx/servex-contrib/transport/grpcx"
)

func TestAppHealthAndShutdown(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	a, srv := newApp("127.0.0.1:0")
	if err := a.Check(ctx); err == nil {
		t.Fatal("app ready before start")
	}
	if err := a.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stop, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := a.Stop(stop); err != nil {
			t.Error(err)
		}
	})
	client, err := grpcx.NewClient(ctx, srv.URL().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	reply, err := grpc_health_v1.NewHealthClient(client.Conn()).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil || reply.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("health=%v error=%v", reply, err)
	}
	if err := a.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Check(ctx); err == nil {
		t.Fatal("app ready after stop")
	}
	if _, err := grpc_health_v1.NewHealthClient(client.Conn()).Check(ctx, &grpc_health_v1.HealthCheckRequest{}); err == nil {
		t.Fatal("server accepted calls after shutdown")
	}
}
