// Package store implements note persistence with the application's shared GORM DB.
package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/nostosx/servex-examples/application/notes"
)

type NoteRepository struct{ db *gorm.DB }

func NewNoteRepository(db *gorm.DB) *NoteRepository { return &NoteRepository{db: db} }
func (r *NoteRepository) Create(ctx context.Context, note notes.Note) error {
	return r.db.WithContext(ctx).Table("example_notes").Create(&note).Error
}
func (r *NoteRepository) List(ctx context.Context, owner string) ([]notes.Note, error) {
	result := []notes.Note{}
	err := r.db.WithContext(ctx).Table("example_notes").Where("owner_id = ?", owner).Order("id").Limit(100).Find(&result).Error
	return result, err
}
func (r *NoteRepository) Update(ctx context.Context, owner, id, text string) error {
	result := r.db.WithContext(ctx).Table("example_notes").Where("owner_id = ? AND id = ?", owner, id).Update("text", text)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return notes.ErrNotFound
	}
	return nil
}
func (r *NoteRepository) Delete(ctx context.Context, owner, id string) error {
	result := r.db.WithContext(ctx).Table("example_notes").Where("owner_id = ? AND id = ?", owner, id).Delete(&notes.Note{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return notes.ErrNotFound
	}
	return nil
}
