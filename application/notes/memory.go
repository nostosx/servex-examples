package notes

import (
	"context"
	"slices"
	"strings"
	"sync"
)

type MemoryRepository struct {
	mu    sync.Mutex
	notes map[string]Note
}

func NewMemoryRepository() *MemoryRepository { return &MemoryRepository{notes: make(map[string]Note)} }

func (r *MemoryRepository) Create(ctx context.Context, note Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := context.Cause(ctx); err != nil {
		return err
	}
	if len(r.notes) >= 10000 {
		return ErrCapacity
	}
	r.notes[note.ID] = note
	return nil
}
func (r *MemoryRepository) List(ctx context.Context, owner string) ([]Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	result := []Note{}
	for _, note := range r.notes {
		if note.OwnerID == owner {
			result = append(result, note)
		}
	}
	slices.SortFunc(result, func(a, b Note) int { return strings.Compare(a.ID, b.ID) })
	if len(result) > 100 {
		result = result[:100]
	}
	return result, nil
}
func (r *MemoryRepository) Update(ctx context.Context, owner, id, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := context.Cause(ctx); err != nil {
		return err
	}
	note, exists := r.notes[id]
	if !exists || note.OwnerID != owner {
		return ErrNotFound
	}
	note.Text = text
	r.notes[id] = note
	return nil
}
func (r *MemoryRepository) Delete(ctx context.Context, owner, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := context.Cause(ctx); err != nil {
		return err
	}
	if note, exists := r.notes[id]; !exists || note.OwnerID != owner {
		return ErrNotFound
	}
	delete(r.notes, id)
	return nil
}
