// Package rpc maps protocols to the service application handler aggregates.
package rpc

import (
	pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"
	serviceapp "github.com/nostosx/servex-examples/services/user/internal/application"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	commands *serviceapp.Command
	queries  *serviceapp.Query
}

func NewUserServer(commands *serviceapp.Command, queries *serviceapp.Query) *UserServer {
	return &UserServer{commands: commands, queries: queries}
}
