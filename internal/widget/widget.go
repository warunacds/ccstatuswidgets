package widget

import "github.com/warunacds/ccstatuswidgets/internal/protocol"

// Widget is the interface that all status line widgets must implement.
type Widget interface {
	// Name returns the unique identifier for this widget.
	Name() string
	// Render produces the widget's output given the current status line input and config.
	Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error)
}
