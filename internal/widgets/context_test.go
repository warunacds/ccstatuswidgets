package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestContextBarWidget_Name(t *testing.T) {
	w := &ContextBarWidget{}
	if w.Name() != "context-bar" {
		t.Errorf("expected name %q, got %q", "context-bar", w.Name())
	}
}

func TestContextBarWidget_RendersBarWithPercentage(t *testing.T) {
	w := &ContextBarWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			UsedPercentage:      30,
			RemainingPercentage: 70,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "ctx \u2588\u2588\u2588\u2591\u2591\u2591\u2591\u2591\u2591\u2591 30%"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestContextBarWidget_ReturnsNilWhenNoContextData(t *testing.T) {
	w := &ContextBarWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			UsedPercentage:      0,
			RemainingPercentage: 0,
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

func TestContextBarWidget_RespectsBarLengthConfig(t *testing.T) {
	w := &ContextBarWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			UsedPercentage:      50,
			RemainingPercentage: 50,
		},
	}
	cfg := map[string]interface{}{
		"bar_length": float64(20),
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "ctx \u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2591\u2591\u2591\u2591\u2591\u2591\u2591\u2591\u2591\u2591 50%"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestContextBarWidget_ShowPercentageFalse(t *testing.T) {
	w := &ContextBarWidget{}
	input := &protocol.StatusLineInput{
		ContextWindow: protocol.ContextInfo{
			UsedPercentage:      30,
			RemainingPercentage: 70,
		},
	}
	cfg := map[string]interface{}{
		"show_percentage": false,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "ctx \u2588\u2588\u2588\u2591\u2591\u2591\u2591\u2591\u2591\u2591"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestContextBarWidget_ColorChangesBasedOnPercentage(t *testing.T) {
	w := &ContextBarWidget{}
	tests := []struct {
		pct       float64
		wantColor string
	}{
		{30, "green"},
		{60, "yellow"},
		{90, "red"},
	}
	for _, tt := range tests {
		input := &protocol.StatusLineInput{
			ContextWindow: protocol.ContextInfo{
				UsedPercentage:      tt.pct,
				RemainingPercentage: 100 - tt.pct,
			},
		}
		out, err := w.Render(input, nil)
		if err != nil {
			t.Fatalf("unexpected error at pct=%v: %v", tt.pct, err)
		}
		if out == nil {
			t.Fatalf("expected non-nil output at pct=%v", tt.pct)
		}
		if out.Color != tt.wantColor {
			t.Errorf("at pct=%v: expected color %q, got %q", tt.pct, tt.wantColor, out.Color)
		}
	}
}
