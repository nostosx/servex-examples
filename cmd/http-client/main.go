// Command http-client demonstrates the servex HTTP client.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/nostosx/servex/transport/httpx"
)

type getUserRequest struct {
	ID int `json:"id"`
}

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func run() error {
	target := flag.String("target", "http://127.0.0.1:8080", "server URL")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	reply, err := fetchUser(ctx, *target)
	if err != nil {
		return err
	}
	fmt.Println(reply.Name)
	return nil
}

func fetchUser(ctx context.Context, target string) (*user, error) {
	// A "discovery:///user" target with httpx.WithDiscoverer resolves nodes
	// through service discovery instead.
	client, err := httpx.NewClient(ctx, target)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	get := httpx.Call[user](client, "GET", "/v1/users/{id}")
	reply, err := get(ctx, &getUserRequest{ID: 7})
	if err != nil {
		return nil, err // an *errorx.Error for non-2xx replies
	}
	return reply.(*user), nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
