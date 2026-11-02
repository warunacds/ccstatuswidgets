package widgets

import (
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// PRWidget displays the open pull request (or GitLab merge request) for the current branch.
type PRWidget struct{}

func (w *PRWidget) Name() string {
	return "pr"
}

func (w *PRWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if input.PR == nil || input.PR.Number == 0 {
		return nil, nil
	}

	label := "PR"
	if input.PR.Kind == "mr" {
		label = "MR"
	}
	text := fmt.Sprintf("%s #%d", label, input.PR.Number)

	switch input.PR.ReviewState {
	case "approved":
		return &protocol.WidgetOutput{Text: text + " ✓", Color: "green"}, nil
	case "changes_requested":
		return &protocol.WidgetOutput{Text: text + " ✗", Color: "red"}, nil
	case "pending":
		return &protocol.WidgetOutput{Text: text + " ○", Color: "yellow"}, nil
	case "draft":
		return &protocol.WidgetOutput{Text: text + " draft", Color: "dim"}, nil
	default:
		return &protocol.WidgetOutput{Text: text, Color: "cyan"}, nil
	}
}
