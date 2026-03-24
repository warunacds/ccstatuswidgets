package widgets

import (
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// CostWidget displays the session cost in USD.
type CostWidget struct{}

func (w *CostWidget) Name() string {
	return "cost"
}

func (w *CostWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	cost := input.Cost.TotalCostUSD
	if cost == 0 {
		return nil, nil
	}

	var text string
	if input.RateLimits != nil {
		text = fmt.Sprintf("api eq. $%.2f", cost)
	} else {
		text = fmt.Sprintf("$%.2f", cost)
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "dim",
	}, nil
}
