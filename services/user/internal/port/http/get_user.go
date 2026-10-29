// Package http adapts the user query to servex typed HTTP routes.
package http

import (
	"context"

	"github.com/nostosx/servex-contrib/transport/httpx"
	"github.com/nostosx/servex-examples/application/user/query"
)

type GetUserRequest struct {
	ID int `json:"id"`
}
type GetUserResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (*Handler) GetUser(ctx context.Context, r *GetUserRequest) (*GetUserResponse, error) {
	reply, err := query.GetUser(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	return &GetUserResponse{ID: reply.ID, Name: reply.Name}, nil
}

// RegisterGetUser mounts the explicitly selected HTTP query.
func RegisterGetUser(server *httpx.Server, handler *Handler) {
	httpx.RouteFunc(server, "GET /v1/users/{id}", handler.GetUser)
}
