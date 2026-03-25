package renderer

import (
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// WidgetResult pairs a widget's output with its name.
// A nil Output means the widget was omitted or timed out.
type WidgetResult struct {
	Output *protocol.WidgetOutput
	Name   string
}

// Render takes ordered widget results grouped by line and produces
// the final colored string for stdout. Widgets on the same line are
// joined with the configured separator (default " "). Lines are
// separated by \n. Empty lines (all widgets nil) are skipped.
// No trailing newline on the last line.
//
// If cfg is nil, defaults are used: separator=" ", no per-widget styling
// (falls back to WidgetOutput.Color for basic ANSI coloring).
func Render(lines [][]WidgetResult, cfg *config.Config) string {
	separator := " "
	var widgetsCfg map[string]map[string]interface{}

	if cfg != nil {
		if cfg.Separator != "" {
			separator = cfg.Separator
		}
		widgetsCfg = cfg.Widgets
	}

	var rendered []string

	for _, line := range lines {
		var parts []string
		for _, wr := range line {
			if wr.Output == nil {
				continue
			}
			text := wr.Output.Text

			// Look up per-widget config for styling.
			var wCfg map[string]interface{}
			if widgetsCfg != nil {
				wCfg = widgetsCfg[wr.Name]
			}

			style := styleFromConfig(wCfg, wr.Output.Color)

			if style.HasStyle() {
				prefix := style.Prefix()
				if prefix != "" {
					text = prefix + text + AnsiReset
				}
			}

			parts = append(parts, text)
		}
		if len(parts) == 0 {
			continue
		}
		rendered = append(rendered, strings.Join(parts, separator))
	}

	return strings.Join(rendered, "\n")
}

// styleFromConfig extracts a WidgetStyle from a widget's config map,
// falling back to the widget's default color if no custom fg is set.
func styleFromConfig(widgetCfg map[string]interface{}, defaultColor string) WidgetStyle {
	style := WidgetStyle{}
	if widgetCfg != nil {
		if v, ok := widgetCfg["fg"].(string); ok {
			style.Fg = v
		}
		if v, ok := widgetCfg["bg"].(string); ok {
			style.Bg = v
		}
		if v, ok := widgetCfg["bold"].(bool); ok {
			style.Bold = v
		}
		if v, ok := widgetCfg["dim"].(bool); ok {
			style.Dim = v
		}
		if v, ok := widgetCfg["italic"].(bool); ok {
			style.Italic = v
		}
		if v, ok := widgetCfg["underline"].(bool); ok {
			style.Underline = v
		}
	}
	// Fall back to widget default color if no custom fg.
	if style.Fg == "" && defaultColor != "" {
		style.Fg = defaultColor
	}
	return style
}
