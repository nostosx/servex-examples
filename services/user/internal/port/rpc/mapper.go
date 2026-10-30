package rpc

import (
	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	"github.com/nostosx/servex-examples/application/user/query"
)

func mapGetUserRequest(request *pb.GetUserRequest) query.GetUserQuery {
	return query.GetUserQuery{ID: int(request.GetId())}
}
func mapGetUserResponse(result query.GetUserResult) *pb.GetUserResponse {
	return &pb.GetUserResponse{Id: int32(result.ID), Name: result.Name}
}
