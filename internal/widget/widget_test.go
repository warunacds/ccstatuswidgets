package widget_test

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

type mockWidget struct{}

func (m *mockWidget) Name() string { return "mock" }
func (m *mockWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	return &protocol.WidgetOutput{Text: "hello", Color: "green"}, nil
}

func TestWidgetInterface(t *testing.T) {
	var w widget.Widget = &mockWidget{}
	if w.Name() != "mock" {
		t.Fatalf("expected mock, got %s", w.Name())
	}
	out, err := w.Render(&protocol.StatusLineInput{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "hello" || out.Color != "green" {
		t.Fatalf("unexpected output: %+v", out)
	}
}
