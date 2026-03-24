package widgets

import (
	"fmt"
	"testing"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestUsage5hWidget_Name(t *testing.T) {
	w := &Usage5hWidget{}
	if w.Name() != "usage-5h" {
		t.Errorf("expected name %q, got %q", "usage-5h", w.Name())
	}
}

func TestUsage7dWidget_Name(t *testing.T) {
	w := &Usage7dWidget{}
	if w.Name() != "usage-7d" {
		t.Errorf("expected name %q, got %q", "usage-7d", w.Name())
	}
}

func TestUsage5hWidget_RendersBarWithPercentageAndPace(t *testing.T) {
	w := &Usage5hWidget{}
	// ResetsAt is 2.5 hours from now (half of 5h window elapsed)
	resetsAt := time.Now().Add(2*time.Hour + 30*time.Minute).Unix()
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{
				UsedPercentage: 22,
				ResetsAt:       resetsAt,
			},
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// With 50% elapsed and 22% used, pace = 22 - 50 = -28 (headroom)
	// Bar: 2 filled + 8 empty (22% of 10)
	bar := BuildBar(22, 10)
	resetStr := formatRelativeTime(resetsAt)
	expected := fmt.Sprintf("5h %s 22%% -28%% \u21bb%s", bar, resetStr)
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestUsage7dWidget_RendersBar(t *testing.T) {
	w := &Usage7dWidget{}
	// ResetsAt is 3.5 days from now (half of 7d window elapsed)
	resetsAt := time.Now().Add(3*24*time.Hour + 12*time.Hour).Unix()
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			SevenDay: &protocol.RateLimit{
				UsedPercentage: 15,
				ResetsAt:       resetsAt,
			},
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// With 50% elapsed and 15% used, pace = 15 - 50 = -35 (headroom)
	bar := BuildBar(15, 10)
	resetStr := formatRelativeTime(resetsAt)
	expected := fmt.Sprintf("7d %s 15%% -35%% \u21bb%s", bar, resetStr)
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestUsage5hWidget_ReturnsNilWhenRateLimitsNil(t *testing.T) {
	w := &Usage5hWidget{}
	input := &protocol.StatusLineInput{
		RateLimits: nil,
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestUsage7dWidget_ReturnsNilWhenRateLimitsNil(t *testing.T) {
	w := &Usage7dWidget{}
	input := &protocol.StatusLineInput{
		RateLimits: nil,
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestUsage5hWidget_ReturnsNilWhenSpecificRateLimitNil(t *testing.T) {
	w := &Usage5hWidget{}
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: nil,
			SevenDay: &protocol.RateLimit{UsedPercentage: 50, ResetsAt: time.Now().Add(time.Hour).Unix()},
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

func TestUsage7dWidget_ReturnsNilWhenSpecificRateLimitNil(t *testing.T) {
	w := &Usage7dWidget{}
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{UsedPercentage: 50, ResetsAt: time.Now().Add(time.Hour).Unix()},
			SevenDay: nil,
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

func TestUsage5hWidget_ShowPaceFalse(t *testing.T) {
	w := &Usage5hWidget{}
	resetsAt := time.Now().Add(2*time.Hour + 30*time.Minute).Unix()
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{
				UsedPercentage: 22,
				ResetsAt:       resetsAt,
			},
		},
	}
	cfg := map[string]interface{}{
		"show_pace": false,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	bar := BuildBar(22, 10)
	resetStr := formatRelativeTime(resetsAt)
	expected := fmt.Sprintf("5h %s 22%% \u21bb%s", bar, resetStr)
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestUsage5hWidget_ShowPercentageFalse(t *testing.T) {
	w := &Usage5hWidget{}
	resetsAt := time.Now().Add(2*time.Hour + 30*time.Minute).Unix()
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{
				UsedPercentage: 22,
				ResetsAt:       resetsAt,
			},
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
	bar := BuildBar(22, 10)
	resetStr := formatRelativeTime(resetsAt)
	// pace = 22 - 50 = -28
	expected := fmt.Sprintf("5h %s -28%% \u21bb%s", bar, resetStr)
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestUsagePace_PositiveWhenBurningFast(t *testing.T) {
	w := &Usage5hWidget{}
	// ResetsAt is 4 hours from now (only 1 hour of 5h elapsed = 20% elapsed)
	resetsAt := time.Now().Add(4 * time.Hour).Unix()
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{
				UsedPercentage: 60,
				ResetsAt:       resetsAt,
			},
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// elapsed = 20%, used = 60%, pace = 60 - 20 = +40 (burning fast)
	bar := BuildBar(60, 10)
	resetStr := formatRelativeTime(resetsAt)
	expected := fmt.Sprintf("5h %s 60%% +40%% \u21bb%s", bar, resetStr)
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestUsagePace_NegativeWhenHeadroom(t *testing.T) {
	w := &Usage5hWidget{}
	// ResetsAt is 1 hour from now (4 hours of 5h elapsed = 80% elapsed)
	resetsAt := time.Now().Add(1 * time.Hour).Unix()
	input := &protocol.StatusLineInput{
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{
				UsedPercentage: 30,
				ResetsAt:       resetsAt,
			},
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// elapsed = 80%, used = 30%, pace = 30 - 80 = -50 (headroom)
	bar := BuildBar(30, 10)
	resetStr := formatRelativeTime(resetsAt)
	expected := fmt.Sprintf("5h %s 30%% -50%% \u21bb%s", bar, resetStr)
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestFormatRelativeTime_Hours(t *testing.T) {
	target := time.Now().Add(2*time.Hour + 30*time.Minute).Unix()
	result := formatRelativeTime(target)
	if result != "2h30m" {
		t.Errorf("expected %q, got %q", "2h30m", result)
	}
}

func TestFormatRelativeTime_Days(t *testing.T) {
	target := time.Now().Add(3*24*time.Hour + 12*time.Hour).Unix()
	result := formatRelativeTime(target)
	if result != "3d12h" {
		t.Errorf("expected %q, got %q", "3d12h", result)
	}
}
