// Package notes contains the application use cases, independent of SDKs and DI.
package notes

import (
	"context"
	"errors"
	"strings"
	"time"
	"uuid"
)

var (
	ErrInvalid  = errors.New("invalid note")
	ErrNotFound = errors.New("note not found")
	ErrCapacity = errors.New("note capacity reached")
)

type Note struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"-"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	Create(context.Context, Note) error
	List(context.Context, string) ([]Note, error)
	Update(context.Context, string, string, string) error
	Delete(context.Context, string, string) error
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Create(ctx context.Context, owner, text string) (Note, error) {
	if !valid(owner, text) {
		return Note{}, ErrInvalid
	}
	note := Note{ID: uuid.NewV7().String(), OwnerID: owner, Text: text, CreatedAt: time.Now().UTC()}
	return note, s.repository.Create(ctx, note)
}
func (s *Service) List(ctx context.Context, owner string) ([]Note, error) {
	if owner == "" {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, owner)
}
func (s *Service) Update(ctx context.Context, owner, id, text string) error {
	if !valid(owner, text) || id == "" {
		return ErrInvalid
	}
	return s.repository.Update(ctx, owner, id, text)
}
func (s *Service) Delete(ctx context.Context, owner, id string) error {
	if owner == "" || id == "" {
		return ErrInvalid
	}
	return s.repository.Delete(ctx, owner, id)
}
func valid(owner, text string) bool {
	return owner != "" && strings.TrimSpace(text) != "" && len(text) <= 2048
}
