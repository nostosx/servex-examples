package rpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
)

func TestGetUser(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		id   int32
		code codes.Code
	}{
		{name: "valid", id: 7, code: codes.OK},
		{name: "zero", code: codes.InvalidArgument},
		{name: "negative", id: -1, code: codes.InvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reply, err := new(Handler).GetUser(t.Context(), &pb.GetUserRequest{Id: tt.id})
			if status.Code(err) != tt.code {
				t.Fatalf("reply=%v err=%v", reply, err)
			}
			if tt.code == codes.OK && (reply.GetId() != 7 || reply.GetName() != "alice") {
				t.Fatal(reply)
			}
		})
	}
}

func TestGetUserCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := new(Handler).GetUser(ctx, &pb.GetUserRequest{Id: 7})
	if status.Code(err) != codes.Canceled {
		t.Fatal(err)
	}
}
