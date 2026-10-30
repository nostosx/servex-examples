package query

import (
	"context"

	"github.com/nostosx/servex-examples/domain/user"
	"github.com/nostosx/servex/cqrs"
	"github.com/nostosx/servex/errorx"
)

type GetUserQuery struct{ ID int }
type GetUserResult = user.User

type GetUserHandler struct{}

func NewGetUserHandler() cqrs.Query[GetUserQuery, GetUserResult] { return &GetUserHandler{} }

func (*GetUserHandler) Handle(ctx context.Context, input GetUserQuery) (GetUserResult, error) {
	if err := ctx.Err(); err != nil {
		return user.User{}, err
	}
	if input.ID <= 0 {
		return user.User{}, errorx.New(errorx.InvalidArgument, "INVALID_ID", "id must be positive")
	}
	return user.User{ID: input.ID, Name: "alice"}, nil
}
