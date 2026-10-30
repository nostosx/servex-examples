package notes

import (
	"errors"
	"testing"
)

func TestNotesOwnerIsolation(t *testing.T) {
	service := NewService(NewMemoryRepository())
	created, err := service.Create(t.Context(), "owner-a", "private note")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Update(t.Context(), "owner-b", created.ID, "stolen"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := service.Delete(t.Context(), "owner-b", created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	other, err := service.List(t.Context(), "owner-b")
	if err != nil || len(other) != 0 {
		t.Fatalf("other owner: %v %v", other, err)
	}
	own, err := service.List(t.Context(), "owner-a")
	if err != nil || len(own) != 1 || own[0].Text != "private note" {
		t.Fatalf("own: %v %v", own, err)
	}
	own[0].Text = "mutated response"
	saved, err := service.List(t.Context(), "owner-a")
	if err != nil || saved[0].Text != "private note" {
		t.Fatalf("storage alias: %v %v", saved, err)
	}
	if _, err := service.Create(t.Context(), "owner-a", " "); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
