package http

import (
	"github.com/nostosx/servex-examples/application/user/query"
)

func mapGetUserRequest(request *GetUserRequest) query.GetUserQuery {
	return query.GetUserQuery{ID: request.ID}
}
func mapGetUserResponse(result query.GetUserResult) *GetUserResponse {
	return &GetUserResponse{ID: result.ID, Name: result.Name}
}
