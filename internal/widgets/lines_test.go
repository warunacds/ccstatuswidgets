package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestLinesWidget_Name(t *testing.T) {
	w := &LinesWidget{}
	if w.Name() != "lines-changed" {
		t.Errorf("expected name %q, got %q", "lines-changed", w.Name())
	}
}

func TestLinesWidget_RendersGreenAndRedANSI(t *testing.T) {
	w := &LinesWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalLinesAdded:   42,
			TotalLinesRemoved: 7,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "\033[0;32m+42\033[0m \033[0;31m-7\033[0m"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "" {
		t.Errorf("expected empty color, got %q", out.Color)
	}
}

func TestLinesWidget_ReturnsNilWhenBothZero(t *testing.T) {
	w := &LinesWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalLinesAdded:   0,
			TotalLinesRemoved: 0,
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

func TestLinesWidget_OnlyAdditions(t *testing.T) {
	w := &LinesWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalLinesAdded:   15,
			TotalLinesRemoved: 0,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "\033[0;32m+15\033[0m \033[0;31m-0\033[0m"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestLinesWidget_OnlyRemovals(t *testing.T) {
	w := &LinesWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalLinesAdded:   0,
			TotalLinesRemoved: 23,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "\033[0;32m+0\033[0m \033[0;31m-23\033[0m"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}
