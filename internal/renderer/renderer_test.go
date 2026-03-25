package renderer

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestSingleWidget(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "hello"}, Name: "greeting"},
		},
	}
	got := Render(results, nil)
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
	got := Render(results, nil)
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
	got := Render(results, nil)
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
	got := Render(results, nil)
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
	got := Render(results, nil)
	want := "\033[32mok\033[0m"
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
	got := Render(results, nil)
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
	got := Render(results, nil)
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
	got := Render(results, nil)
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
	got := Render(results, nil)
	want := "plain"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmptyInput(t *testing.T) {
	got := Render(nil, nil)
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
	got = Render([][]WidgetResult{}, nil)
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
	got := Render(results, nil)
	want := "\033[31merr\033[0m \033[33mwarn\033[0m \033[32mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- New tests for per-widget styling and custom separators ---

func TestCustomFgFromConfig(t *testing.T) {
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{
			"status": {"fg": "cyan"},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	want := "\033[36mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCustomFgAndBgFromConfig(t *testing.T) {
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{
			"status": {"fg": "white", "bg": "red"},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "alert"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	want := "\033[37;41malert\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBoldFromConfig(t *testing.T) {
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{
			"status": {"bold": true},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "important"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	want := "\033[1mimportant\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBoldAndFgFromConfig(t *testing.T) {
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{
			"status": {"fg": "green", "bold": true},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	// Bold (1) comes before fg (32) per WidgetStyle.Prefix() ordering.
	want := "\033[1;32mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNoCustomStyleUsesDefaultColor(t *testing.T) {
	// Config exists but has no styling for this widget — should use WidgetOutput.Color.
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{
			"other": {"fg": "red"},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok", Color: "green"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	want := "\033[32mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCustomFgOverridesDefaultColor(t *testing.T) {
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{
			"status": {"fg": "cyan"},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok", Color: "green"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	// Custom fg (cyan=36) overrides default color (green).
	want := "\033[36mok\033[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCustomSeparator(t *testing.T) {
	cfg := &config.Config{
		Separator: " | ",
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "left"}, Name: "a"},
			{Output: &protocol.WidgetOutput{Text: "right"}, Name: "b"},
		},
	}
	got := Render(results, cfg)
	want := "left | right"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNilConfigBackwardCompat(t *testing.T) {
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok", Color: "green"}, Name: "status"},
			{Output: &protocol.WidgetOutput{Text: "hi"}, Name: "greet"},
		},
	}
	got := Render(results, nil)
	want := "\033[32mok\033[0m hi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRawANSIPassedThroughWithNoCustomStyle(t *testing.T) {
	// Widget text already has raw ANSI, no custom style in config — pass through as-is.
	raw := "\033[1;31m+42\033[0m \033[32m-10\033[0m"
	cfg := &config.Config{
		Widgets: map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: raw}, Name: "lines-changed"},
		},
	}
	got := Render(results, cfg)
	if got != raw {
		t.Errorf("got %q, want %q", got, raw)
	}
}

func TestStyleFromConfigHelper(t *testing.T) {
	tests := []struct {
		name         string
		widgetCfg    map[string]interface{}
		defaultColor string
		want         WidgetStyle
	}{
		{
			name:         "nil config uses default color",
			widgetCfg:    nil,
			defaultColor: "red",
			want:         WidgetStyle{Fg: "red"},
		},
		{
			name:         "empty config uses default color",
			widgetCfg:    map[string]interface{}{},
			defaultColor: "blue",
			want:         WidgetStyle{Fg: "blue"},
		},
		{
			name:         "custom fg overrides default",
			widgetCfg:    map[string]interface{}{"fg": "cyan"},
			defaultColor: "red",
			want:         WidgetStyle{Fg: "cyan"},
		},
		{
			name:         "all fields",
			widgetCfg:    map[string]interface{}{"fg": "white", "bg": "blue", "bold": true, "dim": true, "italic": true, "underline": true},
			defaultColor: "",
			want:         WidgetStyle{Fg: "white", Bg: "blue", Bold: true, Dim: true, Italic: true, Underline: true},
		},
		{
			name:         "no default no config",
			widgetCfg:    nil,
			defaultColor: "",
			want:         WidgetStyle{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := styleFromConfig(tt.widgetCfg, tt.defaultColor)
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
