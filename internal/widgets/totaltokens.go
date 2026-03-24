package widgets

import (
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// TotalTokensWidget displays the combined input + output token count.
type TotalTokensWidget struct{}

func (w *TotalTokensWidget) Name() string {
	return "total-tokens"
}

func (w *TotalTokensWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	total := input.ContextWindow.TotalInputTokens + input.ContextWindow.TotalOutputTokens
	if total == 0 {
		return nil, nil
	}

	return &protocol.WidgetOutput{
		Text:  fmt.Sprintf("%s tokens", compactTokenCount(total)),
		Color: "dim",
	}, nil
}
