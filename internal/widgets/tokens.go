package widgets

import (
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// TokensWidget displays input and output token counts in compact format.
type TokensWidget struct{}

func (w *TokensWidget) Name() string {
	return "tokens"
}

func (w *TokensWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	inTok := input.ContextWindow.TotalInputTokens
	outTok := input.ContextWindow.TotalOutputTokens

	if inTok == 0 && outTok == 0 {
		return nil, nil
	}

	text := fmt.Sprintf("↑%s ↓%s", formatTokenCount(inTok), formatTokenCount(outTok))

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "dim",
	}, nil
}

// formatTokenCount formats a token count as a compact string.
// <1000 → as-is, 1000-999999 → X.Xk, 1000000+ → X.Xm
func formatTokenCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fm", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
