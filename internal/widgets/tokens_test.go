package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestTokensWidget_Name(t *testing.T) {
	w := &TokensWidget{}
	if got := w.Name(); got != "tokens" {
		t.Errorf("Name() = %q, want %q", got, "tokens")
	}
}

func TestTokensWidget_SmallNumbers(t *testing.T) {
	w := &TokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  342,
			TotalOutputTokens: 125,
		},
	}
	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if out == nil {
		t.Fatal("Render() returned nil")
	}
	want := "↑342 ↓125"
	if out.Text != want {
		t.Errorf("Text = %q, want %q", out.Text, want)
	}
}

func TestTokensWidget_Thousands(t *testing.T) {
	w := &TokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  12400,
			TotalOutputTokens: 8200,
		},
	}
	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if out == nil {
		t.Fatal("Render() returned nil")
	}
	want := "↑12.4k ↓8.2k"
	if out.Text != want {
		t.Errorf("Text = %q, want %q", out.Text, want)
	}
}

func TestTokensWidget_Millions(t *testing.T) {
	w := &TokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  1200000,
			TotalOutputTokens: 500000,
		},
	}
	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if out == nil {
		t.Fatal("Render() returned nil")
	}
	want := "↑1.2m ↓500.0k"
	if out.Text != want {
		t.Errorf("Text = %q, want %q", out.Text, want)
	}
}

func TestTokensWidget_BothZero(t *testing.T) {
	w := &TokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  0,
			TotalOutputTokens: 0,
		},
	}
	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if out != nil {
		t.Errorf("Render() = %v, want nil", out)
	}
}

func TestTokensWidget_Color(t *testing.T) {
	w := &TokensWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			TotalInputTokens:  1000,
			TotalOutputTokens: 500,
		},
	}
	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if out == nil {
		t.Fatal("Render() returned nil")
	}
	if out.Color != "dim" {
		t.Errorf("Color = %q, want %q", out.Color, "dim")
	}
}
