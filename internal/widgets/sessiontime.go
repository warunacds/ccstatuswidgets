package widgets

import (
	"fmt"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// SessionTimeWidget displays the session's wall-clock duration as reported by
// Claude Code (cost.total_duration_ms), which accumulates across resumes.
type SessionTimeWidget struct{}

func (w *SessionTimeWidget) Name() string {
	return "session-time"
}

func (w *SessionTimeWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	// Zero means the field is missing (older Claude Code); hide rather than show "0m" forever.
	if input.Cost.TotalDurationMs <= 0 {
		return nil, nil
	}
	elapsed := time.Duration(input.Cost.TotalDurationMs) * time.Millisecond
	totalMinutes := int(elapsed.Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	var text string
	if hours == 0 {
		text = fmt.Sprintf("⏱ %dm", minutes)
	} else if minutes == 0 {
		text = fmt.Sprintf("⏱ %dh", hours)
	} else {
		text = fmt.Sprintf("⏱ %dh%dm", hours, minutes)
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "dim",
	}, nil
}
