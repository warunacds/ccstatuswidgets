package widgets

import "github.com/warunacds/ccstatuswidgets/internal/protocol"

// ModelWidget displays the current model's display name.
type ModelWidget struct{}

func (w *ModelWidget) Name() string {
	return "model"
}

func (w *ModelWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if input.Model.DisplayName == "" {
		return nil, nil
	}
	return &protocol.WidgetOutput{
		Text:  input.Model.DisplayName,
		Color: "magenta",
	}, nil
}
