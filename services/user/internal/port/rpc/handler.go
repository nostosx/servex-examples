// Package rpc maps service contracts to application use cases.
package rpc

import pb "github.com/nostosx/servex-examples/api/gen/go/user/v1"

// Handler embeds forward-compatible defaults until business RPCs are implemented.
type Handler struct {
	pb.UnimplementedUserServiceServer
}
