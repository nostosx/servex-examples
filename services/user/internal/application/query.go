package application

import (
	cqrslog "github.com/nostosx/servex-contrib/middleware/logging"
	"github.com/nostosx/servex-examples/application/notes"
	notequery "github.com/nostosx/servex-examples/application/notes/query"
	userquery "github.com/nostosx/servex-examples/application/user/query"
	"github.com/nostosx/servex/cqrs"
	"github.com/nostosx/servex/logger"
)

type Query struct {
	GetUser   cqrs.Query[userquery.GetUserQuery, userquery.GetUserResult]
	ListNotes cqrs.Query[notequery.ListNotesQuery, []notes.Note]
}
type QueryDeps struct {
	GetUser           cqrs.Query[userquery.GetUserQuery, userquery.GetUserResult]
	ListNotes         cqrs.Query[notequery.ListNotesQuery, []notes.Note]
	Logger            logger.Logger
	GetUserMiddleware []cqrs.QueryMiddleware[userquery.GetUserQuery, userquery.GetUserResult]
}

func NewQuery(deps QueryDeps) *Query {
	middleware := append([]cqrs.QueryMiddleware[userquery.GetUserQuery, userquery.GetUserResult]{cqrslog.Query[userquery.GetUserQuery, userquery.GetUserResult](deps.Logger)}, deps.GetUserMiddleware...)
	return &Query{
		GetUser:   cqrs.ChainQuery("user.get_user", deps.GetUser, middleware...),
		ListNotes: cqrs.ChainQuery("notes.list_notes", deps.ListNotes, cqrslog.Query[notequery.ListNotesQuery, []notes.Note](deps.Logger)),
	}
}
