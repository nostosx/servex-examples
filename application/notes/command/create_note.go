package command

import (
	"context"

	"github.com/nostosx/servex-examples/application/notes"
	"github.com/nostosx/servex-examples/application/notes/port"
	"github.com/nostosx/servex/cqrs"
)

type CreateNoteCommand struct{ Owner, Text string }
type CreateNoteHandler struct{ notes port.Creator }

func NewCreateNoteHandler(notes port.Creator) cqrs.Command[CreateNoteCommand, notes.Note] {
	return &CreateNoteHandler{notes: notes}
}
func (h *CreateNoteHandler) Handle(ctx context.Context, in CreateNoteCommand) (notes.Note, error) {
	if err := ctx.Err(); err != nil {
		var zero notes.Note
		return zero, err
	}
	return h.notes.Create(ctx, in.Owner, in.Text)
}
