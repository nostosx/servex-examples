// Command grpc-client checks a server through the standard gRPC health API.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/nostosx/servex-contrib/transport/grpcx"
)

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	target := flag.String("target", "grpc://127.0.0.1:9000", "server URL")
	flag.Parse()
	status, err := checkHealth(ctx, *target)
	if err != nil {
		return err
	}
	log.Println(status)
	return nil
}

func checkHealth(ctx context.Context, target string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	// "discovery:///greeter" with grpcx.WithDiscoverer resolves nodes through
	// service discovery instead.
	client, err := grpcx.NewClient(ctx, target)
	if err != nil {
		return grpc_health_v1.HealthCheckResponse_UNKNOWN, err
	}
	defer client.Close()

	hc := grpc_health_v1.NewHealthClient(client.Conn())
	resp, err := hc.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return grpc_health_v1.HealthCheckResponse_UNKNOWN, err
	}
	status := resp.GetStatus()
	if status != grpc_health_v1.HealthCheckResponse_SERVING {
		return status, fmt.Errorf("health check: server is %s", status)
	}
	return status, nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
