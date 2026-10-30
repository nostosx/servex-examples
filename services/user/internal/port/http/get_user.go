// Package http adapts the user query to servex typed HTTP routes.
package http

import (
	"context"

	"github.com/nostosx/servex-contrib/transport/httpx"
)

type GetUserRequest struct {
	ID int `json:"id"`
}
type GetUserResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *UserServer) GetUser(ctx context.Context, r *GetUserRequest) (*GetUserResponse, error) {
	reply, err := h.queries.GetUser.Handle(ctx, mapGetUserRequest(r))
	if err != nil {
		return nil, err
	}
	return mapGetUserResponse(reply), nil
}

// RegisterGetUser mounts the explicitly selected HTTP query.
func RegisterGetUser(server *httpx.Server, handler *UserServer) {
	httpx.RouteFunc(server, "GET /v1/users/{id}", handler.GetUser)
}
