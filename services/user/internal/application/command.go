package application

import (
	cqrslog "github.com/nostosx/servex-contrib/middleware/logging"
	"github.com/nostosx/servex-examples/application/notes"
	notecommand "github.com/nostosx/servex-examples/application/notes/command"
	"github.com/nostosx/servex/cqrs"
	"github.com/nostosx/servex/logger"
)

type Command struct {
	CreateNote cqrs.Command[notecommand.CreateNoteCommand, notes.Note]
	UpdateNote cqrs.Command[notecommand.UpdateNoteCommand, struct{}]
	DeleteNote cqrs.Command[notecommand.DeleteNoteCommand, struct{}]
}
type CommandDeps struct {
	CreateNote cqrs.Command[notecommand.CreateNoteCommand, notes.Note]
	UpdateNote cqrs.Command[notecommand.UpdateNoteCommand, struct{}]
	DeleteNote cqrs.Command[notecommand.DeleteNoteCommand, struct{}]
	Logger     logger.Logger
}

func NewCommand(deps CommandDeps) *Command {
	return &Command{
		CreateNote: cqrs.ChainCommand("notes.create_note", deps.CreateNote, cqrslog.Command[notecommand.CreateNoteCommand, notes.Note](deps.Logger)),
		UpdateNote: cqrs.ChainCommand("notes.update_note", deps.UpdateNote, cqrslog.Command[notecommand.UpdateNoteCommand, struct{}](deps.Logger)),
		DeleteNote: cqrs.ChainCommand("notes.delete_note", deps.DeleteNote, cqrslog.Command[notecommand.DeleteNoteCommand, struct{}](deps.Logger)),
	}
}
