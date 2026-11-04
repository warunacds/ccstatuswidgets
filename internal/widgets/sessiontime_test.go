package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestSessionTimeWidget_Name(t *testing.T) {
	w := &SessionTimeWidget{}
	if w.Name() != "session-time" {
		t.Errorf("expected name %q, got %q", "session-time", w.Name())
	}
}

func sessionTimeInput(ms int64) *protocol.StatusLineInput {
	return &protocol.StatusLineInput{Cost: protocol.CostInfo{TotalDurationMs: ms}}
}

func TestSessionTimeWidget_FormatsMinutesOnly(t *testing.T) {
	w := &SessionTimeWidget{}
	out, err := w.Render(sessionTimeInput(14*60*1000+30*1000), nil)
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
	out, err := w.Render(sessionTimeInput((1*60+23)*60*1000), nil)
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
	out, err := w.Render(sessionTimeInput(2*60*60*1000), nil)
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

func TestSessionTimeWidget_HidesWhenDurationMissing(t *testing.T) {
	w := &SessionTimeWidget{}
	out, err := w.Render(sessionTimeInput(0), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when total_duration_ms is absent, got %+v", out)
	}
}

func TestSessionTimeWidget_ColorIsDim(t *testing.T) {
	w := &SessionTimeWidget{}
	out, err := w.Render(sessionTimeInput(5*60*1000), nil)
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
