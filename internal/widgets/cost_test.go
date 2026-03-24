package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestCostWidget_Name(t *testing.T) {
	w := &CostWidget{}
	if w.Name() != "cost" {
		t.Errorf("expected name %q, got %q", "cost", w.Name())
	}
}

func TestCostWidget_ShowsAPIEquivalentForMaxPlan(t *testing.T) {
	w := &CostWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalCostUSD: 1.5,
		},
		RateLimits: &protocol.RateLimits{},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "api eq. $1.50" {
		t.Errorf("expected text %q, got %q", "api eq. $1.50", out.Text)
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}

func TestCostWidget_ShowsDollarAmountForAPIKey(t *testing.T) {
	w := &CostWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalCostUSD: 0.75,
		},
		RateLimits: nil,
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "$0.75" {
		t.Errorf("expected text %q, got %q", "$0.75", out.Text)
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}

func TestCostWidget_ReturnsNilWhenCostZero(t *testing.T) {
	w := &CostWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalCostUSD: 0,
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

func TestCostWidget_FormatsToTwoDecimalPlaces(t *testing.T) {
	w := &CostWidget{}
	input := &protocol.StatusLineInput{
		Cost: protocol.CostInfo{
			TotalCostUSD: 12.3456,
		},
		RateLimits: nil,
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "$12.35" {
		t.Errorf("expected text %q, got %q", "$12.35", out.Text)
	}
}
