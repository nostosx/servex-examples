package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	"github.com/nostosx/servex-examples/application/user/query"
	"github.com/nostosx/servex-examples/infrastructure/rpc"
	"github.com/nostosx/servex/errorx"
)

func (*Handler) GetUser(ctx context.Context, r *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	reply, err := query.GetUser(ctx, int(r.GetId()))
	if err != nil {
		if errorx.CodeOf(err) == errorx.InvalidArgument {
			return nil, status.Error(codes.InvalidArgument, "INVALID_ID")
		}
		return nil, rpc.Error(ctx, err)
	}
	return &pb.GetUserResponse{Id: int32(reply.ID), Name: reply.Name}, nil
}
