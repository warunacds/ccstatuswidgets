package renderer

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestSingleWidget(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "hello"}, Name: "greeting"},
		},
	}
	got := Render(results)
	want := "hello"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMultipleWidgetsOneLine(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "left"}, Name: "a"},
			{Output: &protocol.WidgetOutput{Text: "right"}, Name: "b"},
		},
	}
	got := Render(results)
	want := "left right"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMultipleLinesSeparatedByNewline(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "line1"}, Name: "a"},
		},
		{
			{Output: &protocol.WidgetOutput{Text: "line2"}, Name: "b"},
		},
	}
	got := Render(results)
	want := "line1\nline2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmptyLinesSkipped(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "top"}, Name: "a"},
		},
		{
			// All nil outputs — this line should be skipped.
			{Output: nil, Name: "skipped1"},
			{Output: nil, Name: "skipped2"},
		},
		{
			{Output: &protocol.WidgetOutput{Text: "bottom"}, Name: "c"},
		},
	}
	got := Render(results)
	want := "top\nbottom"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestColorWrapsTextInANSI(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok", Color: "green"}, Name: "status"},
		},
	}
	got := Render(results)
	want := "\033[0;32mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRawANSIPassedThrough(t *testing.T) {
	raw := "\033[1;31mBOLD RED\033[0m normal"
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: raw}, Name: "raw"},
		},
	}
	got := Render(results)
	if got != raw {
		t.Errorf("got %q, want %q", got, raw)
	}
}

func TestNoTrailingNewline(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "a"}, Name: "x"},
		},
		{
			{Output: &protocol.WidgetOutput{Text: "b"}, Name: "y"},
		},
	}
	got := Render(results)
	if len(got) == 0 {
		t.Fatal("output is empty")
	}
	if got[len(got)-1] == '\n' {
		t.Error("output has trailing newline")
	}
}

func TestNilOutputWidgetSkipped(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "visible"}, Name: "a"},
			{Output: nil, Name: "gone"},
			{Output: &protocol.WidgetOutput{Text: "also visible"}, Name: "b"},
		},
	}
	got := Render(results)
	want := "visible also visible"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUnknownColorNoWrap(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "plain", Color: "neon"}, Name: "x"},
		},
	}
	got := Render(results)
	want := "plain"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmptyInput(t *testing.T) {
	got := Render(nil)
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
	got = Render([][]WidgetResult{})
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestMultipleColorsOnOneLine(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "err", Color: "red"}, Name: "a"},
			{Output: &protocol.WidgetOutput{Text: "warn", Color: "yellow"}, Name: "b"},
			{Output: &protocol.WidgetOutput{Text: "ok", Color: "green"}, Name: "c"},
		},
	}
	got := Render(results)
	want := "\033[0;31merr\033[0m \033[0;33mwarn\033[0m \033[0;32mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
