//go:build integration

package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/server"
	lifecycle "github.com/nostosx/servex/app"
	"github.com/nostosx/servex/health"
)

func TestUserHTTPAndRPC(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cfg := config.Config{Address: "127.0.0.1:0", GRPCAddress: "127.0.0.1:0", Token: strings.Repeat("a", 32)}
	var app *lifecycle.App
	servers, err := server.New(ctx, cfg, health.CheckerFunc(func(ctx context.Context) error { return app.Check(ctx) }), nil)
	if err != nil {
		t.Fatal(err)
	}
	commands, queries := testApplicationHandlers()
	Register(servers.HTTP, servers.GRPC, commands, queries)
	app = lifecycle.New(lifecycle.WithHooks(servers.HTTP, servers.GRPC))
	if err := app.Check(ctx); err == nil {
		t.Fatal("ready before start")
	}
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stop, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := app.Stop(stop); err != nil {
			t.Error(err)
		}
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, servers.HTTP.URL().String()+"/v1/users/7", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		ID   int
		Name string
	}
	decodeErr := json.NewDecoder(response.Body).Decode(&got)
	closeErr := response.Body.Close()
	if decodeErr != nil || closeErr != nil || response.StatusCode != http.StatusOK || got.ID != 7 || got.Name != "alice" {
		t.Fatalf("user=%+v status=%d decode=%v close=%v", got, response.StatusCode, decodeErr, closeErr)
	}
	conn, err := grpc.NewClient(servers.GRPC.URL().Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewUserServiceClient(conn)
	if _, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 7}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated query: %v", err)
	}
	authorized := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+cfg.Token)
	reply, err := client.GetUser(authorized, &pb.GetUserRequest{Id: 7})
	if err != nil || reply.GetId() != 7 || reply.GetName() != "alice" {
		t.Fatalf("reply=%v err=%v", reply, err)
	}
	health := grpc_health_v1.NewHealthClient(conn)
	ready, err := health.Check(authorized, &grpc_health_v1.HealthCheckRequest{})
	if err != nil || ready.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("health=%v err=%v", ready, err)
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.Check(ctx); err == nil {
		t.Fatal("ready after stop")
	}
	if _, err := client.GetUser(authorized, &pb.GetUserRequest{Id: 7}); err == nil {
		t.Fatal("accepted query after shutdown")
	}
}
