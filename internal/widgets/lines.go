package widgets

import (
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// LinesWidget displays lines added and removed with ANSI color codes.
type LinesWidget struct{}

func (w *LinesWidget) Name() string {
	return "lines-changed"
}

func (w *LinesWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	added := input.Cost.TotalLinesAdded
	removed := input.Cost.TotalLinesRemoved

	if added == 0 && removed == 0 {
		return nil, nil
	}

	text := fmt.Sprintf("\033[0;32m+%d\033[0m \033[0;31m-%d\033[0m", added, removed)
	return &protocol.WidgetOutput{
		Text:  text,
		Color: "",
	}, nil
}
