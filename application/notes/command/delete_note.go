package command

import (
	"context"

	"github.com/nostosx/servex-examples/application/notes/port"
	"github.com/nostosx/servex/cqrs"
)

type DeleteNoteCommand struct{ Owner, ID string }
type DeleteNoteHandler struct{ notes port.Deleter }

func NewDeleteNoteHandler(notes port.Deleter) cqrs.Command[DeleteNoteCommand, struct{}] {
	return &DeleteNoteHandler{notes: notes}
}
func (h *DeleteNoteHandler) Handle(ctx context.Context, in DeleteNoteCommand) (struct{}, error) {
	if err := ctx.Err(); err != nil {
		var zero struct{}
		return zero, err
	}
	return struct{}{}, h.notes.Delete(ctx, in.Owner, in.ID)
}
