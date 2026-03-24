package widgets

import (
	"fmt"
	"math"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// MoonWidget displays the current moon phase.
type MoonWidget struct{}

func (w *MoonWidget) Name() string {
	return "moon"
}

func (w *MoonWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	phase := moonPhaseForDate(time.Now())

	emojis := [8]string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	names := [8]string{
		"new moon", "waxing crescent", "first quarter", "waxing gibbous",
		"full moon", "waning gibbous", "last quarter", "waning crescent",
	}

	return &protocol.WidgetOutput{
		Text:  fmt.Sprintf("%s %s", emojis[phase], names[phase]),
		Color: "dim",
	}, nil
}

// moonPhaseForDate calculates the moon phase index (0-7) for a given date
// using the synodic month algorithm.
//
// Reference: January 6, 2000 18:14 UTC was a known new moon.
// Synodic month: 29.53059 days.
//
// Phase indices:
//
//	0 = new moon, 1 = waxing crescent, 2 = first quarter, 3 = waxing gibbous,
//	4 = full moon, 5 = waning gibbous, 6 = last quarter, 7 = waning crescent
func moonPhaseForDate(t time.Time) int {
	const synodicMonth = 29.53059

	// Known new moon: January 6, 2000 at 18:14 UTC
	ref := time.Date(2000, time.January, 6, 18, 14, 0, 0, time.UTC)

	days := t.Sub(ref).Hours() / 24.0
	cycles := days / synodicMonth
	fraction := cycles - math.Floor(cycles)
	if fraction < 0 {
		fraction += 1.0
	}

	// Map the 0.0-1.0 fraction to 8 phases (0-7)
	phase := int(fraction * 8.0)
	if phase > 7 {
		phase = 7
	}
	return phase
}
