package renderer

import (
	"strings"
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

// --- Powerline mode tests ---

func TestPowerlineContainsArrowSeparator(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "model"}, Name: "model"},
			{Output: &protocol.WidgetOutput{Text: "effort"}, Name: "effort"},
		},
	}
	got := Render(results, cfg)
	if !strings.Contains(got, PowerlineArrow) {
		t.Errorf("powerline output should contain arrow separator %q, got %q", PowerlineArrow, got)
	}
}

func TestPowerlineAssignsPaletteColorsWhenNoBg(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "a"}, Name: "w1"},
			{Output: &protocol.WidgetOutput{Text: "b"}, Name: "w2"},
		},
	}
	got := Render(results, cfg)
	// First widget should use palette[0] (#44475a) as bg.
	bg0 := BgCode(powerlinePalette[0])
	if !strings.Contains(got, bg0) {
		t.Errorf("expected palette bg code %q in output %q", bg0, got)
	}
	// Second widget should use palette[1] (#6272a4) as bg.
	bg1 := BgCode(powerlinePalette[1])
	if !strings.Contains(got, bg1) {
		t.Errorf("expected palette bg code %q in output %q", bg1, got)
	}
}

func TestPowerlineUsesCustomBgWhenSet(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets: map[string]map[string]interface{}{
			"status": {"bg": "#ff0000"},
		},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "ok"}, Name: "status"},
		},
	}
	got := Render(results, cfg)
	customBg := BgCode("#ff0000")
	if !strings.Contains(got, customBg) {
		t.Errorf("expected custom bg code %q in output %q", customBg, got)
	}
	// Should NOT contain default palette[0] bg.
	paletteBg := BgCode(powerlinePalette[0])
	if strings.Contains(got, paletteBg) {
		t.Errorf("should not contain default palette bg %q when custom bg is set, got %q", paletteBg, got)
	}
}

func TestPowerlineDisabledNoArrows(t *testing.T) {
	cfg := &config.Config{
		Powerline: false,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "a"}, Name: "w1"},
			{Output: &protocol.WidgetOutput{Text: "b"}, Name: "w2"},
		},
	}
	got := Render(results, cfg)
	if strings.Contains(got, PowerlineArrow) {
		t.Errorf("powerline disabled: output should not contain arrow %q, got %q", PowerlineArrow, got)
	}
}

func TestPowerlineSingleWidgetClosingArrow(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "solo"}, Name: "only"},
		},
	}
	got := Render(results, cfg)
	// Should contain exactly one arrow (the closing arrow).
	count := strings.Count(got, PowerlineArrow)
	if count != 1 {
		t.Errorf("single widget should have exactly 1 closing arrow, got %d in %q", count, got)
	}
	// The closing arrow should use fg = widget's bg (palette[0]), no bg.
	closingFg := FgCode(powerlinePalette[0])
	// The closing arrow segment should have the fg code.
	if !strings.Contains(got, closingFg) {
		t.Errorf("closing arrow should use fg code %q, got %q", closingFg, got)
	}
}

func TestPowerlineMultipleWidgetsArrowCount(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "a"}, Name: "w1"},
			{Output: &protocol.WidgetOutput{Text: "b"}, Name: "w2"},
			{Output: &protocol.WidgetOutput{Text: "c"}, Name: "w3"},
		},
	}
	got := Render(results, cfg)
	// 3 widgets => 2 inter-widget arrows + 1 closing arrow = 3 arrows total.
	count := strings.Count(got, PowerlineArrow)
	if count != 3 {
		t.Errorf("3 widgets should produce 3 arrows, got %d in %q", count, got)
	}
}

func TestPowerlineSkipsNilWidgets(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "a"}, Name: "w1"},
			{Output: nil, Name: "w2"},
			{Output: &protocol.WidgetOutput{Text: "c"}, Name: "w3"},
		},
	}
	got := Render(results, cfg)
	// Only 2 active widgets => 1 inter-widget arrow + 1 closing arrow = 2 arrows.
	count := strings.Count(got, PowerlineArrow)
	if count != 2 {
		t.Errorf("2 active widgets should produce 2 arrows, got %d in %q", count, got)
	}
}

func TestPowerlineEmptyLineSkipped(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "top"}, Name: "w1"},
		},
		{
			{Output: nil, Name: "empty"},
		},
		{
			{Output: &protocol.WidgetOutput{Text: "bottom"}, Name: "w2"},
		},
	}
	got := Render(results, cfg)
	// Lines should be separated by \n, empty line skipped.
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 rendered lines, got %d: %q", len(lines), got)
	}
}

func TestPowerlineArrowTransitionColors(t *testing.T) {
	cfg := &config.Config{
		Powerline: true,
		Widgets:   map[string]map[string]interface{}{},
	}
	results := [][]WidgetResult{
		{
			{Output: &protocol.WidgetOutput{Text: "a"}, Name: "w1"},
			{Output: &protocol.WidgetOutput{Text: "b"}, Name: "w2"},
		},
	}
	got := Render(results, cfg)
	// The arrow between w1 and w2 should have:
	// fg = palette[0] (w1's bg), bg = palette[1] (w2's bg).
	arrowFg := FgCode(powerlinePalette[0])
	arrowBg := BgCode(powerlinePalette[1])
	expectedArrowPrefix := "\033[" + arrowFg + ";" + arrowBg + "m"
	if !strings.Contains(got, expectedArrowPrefix) {
		t.Errorf("expected arrow transition %q in output %q", expectedArrowPrefix, got)
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
