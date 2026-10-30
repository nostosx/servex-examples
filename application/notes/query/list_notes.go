package query

import (
	"context"

	"github.com/nostosx/servex-examples/application/notes"
	"github.com/nostosx/servex-examples/application/notes/port"
	"github.com/nostosx/servex/cqrs"
)

type ListNotesQuery struct{ Owner string }
type ListNotesHandler struct{ notes port.Reader }

func NewListNotesHandler(notes port.Reader) cqrs.Query[ListNotesQuery, []notes.Note] {
	return &ListNotesHandler{notes: notes}
}
func (h *ListNotesHandler) Handle(ctx context.Context, in ListNotesQuery) ([]notes.Note, error) {
	if err := ctx.Err(); err != nil {
		var zero []notes.Note
		return zero, err
	}
	return h.notes.List(ctx, in.Owner)
}
