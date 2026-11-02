package widgets

import (
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// SessionNameWidget displays the session name, truncated to max_length runes.
type SessionNameWidget struct{}

func (w *SessionNameWidget) Name() string {
	return "session-name"
}

func (w *SessionNameWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if input.SessionName == "" {
		return nil, nil
	}

	maxLength := 32
	if cfg != nil {
		if ml, ok := cfg["max_length"].(float64); ok && ml > 0 {
			maxLength = int(ml)
		}
	}

	// Names come from /rename or an AI-generated title and go straight to the
	// terminal, so drop control characters (C0, DEL, C1) before rendering.
	text := strings.Map(func(r rune) rune {
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			return -1
		}
		return r
	}, input.SessionName)
	if text == "" {
		return nil, nil
	}
	runes := []rune(text)
	if len(runes) > maxLength {
		text = string(runes[:maxLength-1]) + "…"
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "blue",
	}, nil
}
