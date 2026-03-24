package widgets

import (
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// ContextBarWidget displays a bar showing context window usage.
type ContextBarWidget struct{}

func (w *ContextBarWidget) Name() string {
	return "context-bar"
}

func (w *ContextBarWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	pct := input.ContextWindow.UsedPercentage
	remaining := input.ContextWindow.RemainingPercentage

	// Return nil when there is no context data.
	if pct == 0 && remaining == 0 {
		return nil, nil
	}

	barLength := 10
	if cfg != nil {
		if bl, ok := cfg["bar_length"].(float64); ok && bl > 0 {
			barLength = int(bl)
		}
	}

	bar := BuildBar(pct, barLength)
	color := BarColor(pct)

	showPercentage := true
	if cfg != nil {
		if sp, ok := cfg["show_percentage"].(bool); ok {
			showPercentage = sp
		}
	}

	text := "ctx " + bar
	if showPercentage {
		text += fmt.Sprintf(" %d%%", int(pct))
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: color,
	}, nil
}
