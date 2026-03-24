package widgets

import (
	"testing"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestSessionTimeWidget_Name(t *testing.T) {
	w := &SessionTimeWidget{}
	if w.Name() != "session-time" {
		t.Errorf("expected name %q, got %q", "session-time", w.Name())
	}
}

func TestSessionTimeWidget_FormatsMinutesOnly(t *testing.T) {
	w := &SessionTimeWidget{}
	input := &protocol.StatusLineInput{}
	// 14 minutes ago.
	startTime := float64(time.Now().Add(-14 * time.Minute).Unix())
	cfg := map[string]interface{}{
		"start_time": startTime,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "⏱ 14m" {
		t.Errorf("expected %q, got %q", "⏱ 14m", out.Text)
	}
}

func TestSessionTimeWidget_FormatsHoursAndMinutes(t *testing.T) {
	w := &SessionTimeWidget{}
	input := &protocol.StatusLineInput{}
	// 1 hour 23 minutes ago.
	startTime := float64(time.Now().Add(-1*time.Hour - 23*time.Minute).Unix())
	cfg := map[string]interface{}{
		"start_time": startTime,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "⏱ 1h23m" {
		t.Errorf("expected %q, got %q", "⏱ 1h23m", out.Text)
	}
}

func TestSessionTimeWidget_FormatsExactHours(t *testing.T) {
	w := &SessionTimeWidget{}
	input := &protocol.StatusLineInput{}
	// Exactly 2 hours ago.
	startTime := float64(time.Now().Add(-2 * time.Hour).Unix())
	cfg := map[string]interface{}{
		"start_time": startTime,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "⏱ 2h" {
		t.Errorf("expected %q, got %q", "⏱ 2h", out.Text)
	}
}

func TestSessionTimeWidget_ColorIsDim(t *testing.T) {
	w := &SessionTimeWidget{}
	input := &protocol.StatusLineInput{}
	startTime := float64(time.Now().Add(-5 * time.Minute).Unix())
	cfg := map[string]interface{}{
		"start_time": startTime,
	}

	out, err := w.Render(input, cfg)
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

func TestSessionTimeWidget_ReturnsNilWhenStartTimeInvalid(t *testing.T) {
	w := &SessionTimeWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"start_time": "not-a-number",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
