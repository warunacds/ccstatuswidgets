package renderer

import (
	"fmt"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// PowerlineArrow is the Powerline right-pointing arrow separator.
const PowerlineArrow = "\ue0b0"

// powerlinePalette provides default background colors for widgets in
// Powerline mode when no explicit bg is configured.
var powerlinePalette = []string{
	"#44475a", "#6272a4", "#bd93f9", "#50fa7b",
	"#ffb86c", "#ff79c6", "#8be9fd", "#f1fa8c",
}

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
//
// When cfg.Powerline is true, widgets are rendered with background
// colors and Powerline arrow separators between them.
func Render(lines [][]WidgetResult, cfg *config.Config) string {
	if cfg != nil && cfg.Powerline {
		return renderPowerline(lines, cfg)
	}

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

// powerlineWidgetBg returns the background color for a widget in
// Powerline mode. If the widget has a custom bg in its config, that
// is used. Otherwise, a color is auto-assigned from powerlinePalette
// using the widget's position index.
func powerlineWidgetBg(name string, widgetsCfg map[string]map[string]interface{}, index int) string {
	if widgetsCfg != nil {
		if wCfg, ok := widgetsCfg[name]; ok {
			if bg, ok := wCfg["bg"].(string); ok && bg != "" {
				return bg
			}
		}
	}
	return powerlinePalette[index%len(powerlinePalette)]
}

// renderPowerline renders all lines in Powerline mode. Each widget
// gets a background color and adjacent widgets are separated by an
// arrow character whose fg = previous widget's bg and bg = next
// widget's bg, creating the seamless arrow transition effect.
func renderPowerline(lines [][]WidgetResult, cfg *config.Config) string {
	var widgetsCfg map[string]map[string]interface{}
	if cfg != nil {
		widgetsCfg = cfg.Widgets
	}

	var rendered []string

	for _, line := range lines {
		// Filter to only non-nil widgets.
		var active []WidgetResult
		for _, wr := range line {
			if wr.Output != nil {
				active = append(active, wr)
			}
		}
		if len(active) == 0 {
			continue
		}

		// Determine bg colors for each active widget.
		bgColors := make([]string, len(active))
		for i, wr := range active {
			bgColors[i] = powerlineWidgetBg(wr.Name, widgetsCfg, i)
		}

		var buf strings.Builder
		for i, wr := range active {
			bg := bgColors[i]

			// Build the style for this widget's text segment.
			var wCfg map[string]interface{}
			if widgetsCfg != nil {
				wCfg = widgetsCfg[wr.Name]
			}
			style := styleFromConfig(wCfg, wr.Output.Color)
			// Force bg to the powerline bg color.
			style.Bg = bg

			// Render the text segment: " text "
			prefix := style.Prefix()
			buf.WriteString(prefix)
			buf.WriteString(" ")
			buf.WriteString(wr.Output.Text)
			buf.WriteString(" ")
			buf.WriteString(AnsiReset)

			// Render the arrow separator.
			if i < len(active)-1 {
				// Arrow: fg = this widget's bg, bg = next widget's bg.
				nextBg := bgColors[i+1]
				arrowFg := FgCode(bg)
				arrowBg := BgCode(nextBg)
				var arrowParts []string
				if arrowFg != "" {
					arrowParts = append(arrowParts, arrowFg)
				}
				if arrowBg != "" {
					arrowParts = append(arrowParts, arrowBg)
				}
				if len(arrowParts) > 0 {
					buf.WriteString(fmt.Sprintf("\033[%sm", strings.Join(arrowParts, ";")))
				}
				buf.WriteString(PowerlineArrow)
				buf.WriteString(AnsiReset)
			} else {
				// Closing arrow: fg = last widget's bg, no bg.
				arrowFg := FgCode(bg)
				if arrowFg != "" {
					buf.WriteString(fmt.Sprintf("\033[%sm", arrowFg))
				}
				buf.WriteString(PowerlineArrow)
				buf.WriteString(AnsiReset)
			}
		}

		rendered = append(rendered, buf.String())
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
