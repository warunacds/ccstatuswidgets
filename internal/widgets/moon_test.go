package widgets

import (
	"strings"
	"testing"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestMoonWidget_Name(t *testing.T) {
	w := &MoonWidget{}
	if w.Name() != "moon" {
		t.Errorf("expected name %q, got %q", "moon", w.Name())
	}
}

func TestMoonWidget_RenderReturnsNonNilWithColorDim(t *testing.T) {
	w := &MoonWidget{}
	input := &protocol.StatusLineInput{}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output, moon is always visible")
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}

func TestMoonWidget_OutputContainsMoonEmojiAndPhaseName(t *testing.T) {
	w := &MoonWidget{}
	input := &protocol.StatusLineInput{}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	moonEmojis := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	phaseNames := []string{
		"new moon", "waxing crescent", "first quarter", "waxing gibbous",
		"full moon", "waning gibbous", "last quarter", "waning crescent",
	}

	hasEmoji := false
	for _, e := range moonEmojis {
		if strings.Contains(out.Text, e) {
			hasEmoji = true
			break
		}
	}
	if !hasEmoji {
		t.Errorf("output %q does not contain a moon emoji", out.Text)
	}

	hasPhase := false
	for _, p := range phaseNames {
		if strings.Contains(out.Text, p) {
			hasPhase = true
			break
		}
	}
	if !hasPhase {
		t.Errorf("output %q does not contain a phase name", out.Text)
	}
}

func TestMoonPhaseForDate_KnownNewMoon(t *testing.T) {
	// January 6, 2000 was a known new moon
	newMoon := time.Date(2000, time.January, 6, 18, 14, 0, 0, time.UTC)
	phase := moonPhaseForDate(newMoon)
	if phase != 0 {
		t.Errorf("expected phase 0 (new moon) for Jan 6 2000, got %d", phase)
	}
}

func TestMoonPhaseForDate_KnownFullMoon(t *testing.T) {
	// Half synodic month (14.765295 days) after reference new moon (Jan 6, 18:14 UTC)
	// yields the calculated full moon point: approximately Jan 21, 2000 13:00 UTC.
	fullMoon := time.Date(2000, time.January, 21, 13, 0, 0, 0, time.UTC)
	phase := moonPhaseForDate(fullMoon)
	if phase != 4 {
		t.Errorf("expected phase 4 (full moon) for Jan 21 2000 13:00 UTC, got %d", phase)
	}
}

func TestMoonPhaseForDate_AlwaysReturns0To7(t *testing.T) {
	// Test across a wide range of dates
	start := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 365; i++ {
		d := start.AddDate(0, 0, i)
		phase := moonPhaseForDate(d)
		if phase < 0 || phase > 7 {
			t.Errorf("phase %d out of range [0,7] for date %s", phase, d.Format("2006-01-02"))
		}
	}
}
