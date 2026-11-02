package widgets

import "github.com/warunacds/ccstatuswidgets/internal/protocol"

// FastModeWidget displays an indicator when Claude Code fast mode is on.
type FastModeWidget struct{}

func (w *FastModeWidget) Name() string {
	return "fast-mode"
}

func (w *FastModeWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if !input.FastMode {
		return nil, nil
	}
	return &protocol.WidgetOutput{
		Text:  "⚡ fast",
		Color: "yellow",
	}, nil
}
