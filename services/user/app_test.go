package user

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"

	"github.com/nostosx/servex-contrib/transport/httpx"

	"github.com/nostosx/servex-examples/infrastructure/config"
)

func TestRunRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := Run(ctx, config.Config{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

type registry struct {
	description    *grpc.ServiceDesc
	implementation any
}

func (r *registry) RegisterService(description *grpc.ServiceDesc, implementation any) {
	r.description = description
	r.implementation = implementation
}
func TestRegister(t *testing.T) {
	var registrar registry
	Register(httpx.NewServer(), &registrar)
	if registrar.description == nil || registrar.description.ServiceName != "user.v1.UserService" || registrar.implementation == nil {
		t.Fatalf("service not registered: %+v", registrar)
	}
}
