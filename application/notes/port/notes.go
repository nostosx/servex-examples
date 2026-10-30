// Package port declares note capabilities consumed by individual use cases.
package port

import (
	"context"

	"github.com/nostosx/servex-examples/application/notes"
)

type Creator interface {
	Create(context.Context, string, string) (notes.Note, error)
}
type Reader interface {
	List(context.Context, string) ([]notes.Note, error)
}
type Updater interface {
	Update(context.Context, string, string, string) error
}
type Deleter interface {
	Delete(context.Context, string, string) error
}
