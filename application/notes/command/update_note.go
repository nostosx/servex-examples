package command

import (
	"context"

	"github.com/nostosx/servex-examples/application/notes/port"
	"github.com/nostosx/servex/cqrs"
)

type UpdateNoteCommand struct{ Owner, ID, Text string }
type UpdateNoteHandler struct{ notes port.Updater }

func NewUpdateNoteHandler(notes port.Updater) cqrs.Command[UpdateNoteCommand, struct{}] {
	return &UpdateNoteHandler{notes: notes}
}
func (h *UpdateNoteHandler) Handle(ctx context.Context, in UpdateNoteCommand) (struct{}, error) {
	if err := ctx.Err(); err != nil {
		var zero struct{}
		return zero, err
	}
	return struct{}{}, h.notes.Update(ctx, in.Owner, in.ID, in.Text)
}
