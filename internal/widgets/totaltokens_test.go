package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestTotalTokensWidget_Name(t *testing.T) {
	w := &TotalTokensWidget{}
	if w.Name() != "total-tokens" {
		t.Errorf("expected name %q, got %q", "total-tokens", w.Name())
	}
}

func TestTotalTokensWidget_FormatsSmallNumbers(t *testing.T) {
	w := &TotalTokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  200,
			TotalOutputTokens: 142,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "342 tokens" {
		t.Errorf("expected text %q, got %q", "342 tokens", out.Text)
	}
}

func TestTotalTokensWidget_FormatsThousands(t *testing.T) {
	w := &TotalTokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  15000,
			TotalOutputTokens: 5600,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "20.6k tokens" {
		t.Errorf("expected text %q, got %q", "20.6k tokens", out.Text)
	}
}

func TestTotalTokensWidget_FormatsMillions(t *testing.T) {
	w := &TotalTokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  800000,
			TotalOutputTokens: 400000,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "1.2m tokens" {
		t.Errorf("expected text %q, got %q", "1.2m tokens", out.Text)
	}
}

func TestTotalTokensWidget_ReturnsNilWhenZero(t *testing.T) {
	w := &TotalTokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  0,
			TotalOutputTokens: 0,
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

func TestTotalTokensWidget_ColorIsDim(t *testing.T) {
	w := &TotalTokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  500,
			TotalOutputTokens: 500,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}
