package cli_test

import "github.com/warunacds/ccstatuswidgets/internal/protocol"

// fakeWidget is a test helper that satisfies the widget.Widget interface.
type fakeWidget struct {
	name string
}

func (f *fakeWidget) Name() string { return f.name }

func (f *fakeWidget) Render(_ *protocol.StatusLineInput, _ map[string]interface{}) (*protocol.WidgetOutput, error) {
	return &protocol.WidgetOutput{Text: f.name}, nil
}
