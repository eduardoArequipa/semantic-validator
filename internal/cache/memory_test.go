package cache

import (
	"testing"

	"github.com/eduardoArequipa/semantic-validator/internal/validation"
)

func TestMemoryGetSet(t *testing.T) {
	memory := NewMemory()
	want := validation.Result{Confidence: .98, Status: "valid"}
	memory.Set("key", want)
	got, found := memory.Get("key")
	if !found || got.Confidence != want.Confidence || got.Status != want.Status {
		t.Fatalf("got=%+v found=%v", got, found)
	}
}
