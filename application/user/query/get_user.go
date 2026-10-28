package query

import (
	"context"

	"github.com/nostosx/servex-examples/domain/user"
	"github.com/nostosx/servex/errorx"
)

func GetUser(ctx context.Context, id int) (user.User, error) {
	if err := ctx.Err(); err != nil {
		return user.User{}, err
	}
	if id <= 0 {
		return user.User{}, errorx.New(errorx.InvalidArgument, "INVALID_ID", "id must be positive")
	}
	return user.User{ID: id, Name: "alice"}, nil
}
