package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestModelWidget_Name(t *testing.T) {
	w := &ModelWidget{}
	if w.Name() != "model" {
		t.Errorf("expected name %q, got %q", "model", w.Name())
	}
}

func TestModelWidget_ReturnsDisplayNameInMagenta(t *testing.T) {
	w := &ModelWidget{}
	input := &protocol.StatusLineInput{
		Model: protocol.ModelInfo{
			ID:          "claude-opus-4-6",
			DisplayName: "Opus 4.6",
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "Opus 4.6" {
		t.Errorf("expected text %q, got %q", "Opus 4.6", out.Text)
	}
	if out.Color != "magenta" {
		t.Errorf("expected color %q, got %q", "magenta", out.Color)
	}
}

func TestModelWidget_ReturnsNilWhenEmpty(t *testing.T) {
	w := &ModelWidget{}
	input := &protocol.StatusLineInput{
		Model: protocol.ModelInfo{
			ID:          "claude-opus-4-6",
			DisplayName: "",
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
