package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestFastModeWidget_Name(t *testing.T) {
	w := &FastModeWidget{}
	if w.Name() != "fast-mode" {
		t.Errorf("expected name %q, got %q", "fast-mode", w.Name())
	}
}

func TestFastModeWidget_ReturnsNilWhenOff(t *testing.T) {
	w := &FastModeWidget{}
	input := &protocol.StatusLineInput{FastMode: false}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestFastModeWidget_ShownWhenOn(t *testing.T) {
	w := &FastModeWidget{}
	input := &protocol.StatusLineInput{FastMode: true}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "⚡ fast" {
		t.Errorf("expected text %q, got %q", "⚡ fast", out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}
