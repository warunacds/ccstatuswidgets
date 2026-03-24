package renderer

import (
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// WidgetResult pairs a widget's output with its name.
// A nil Output means the widget was omitted or timed out.
type WidgetResult struct {
	Output *protocol.WidgetOutput
	Name   string
}

// colorMap maps color names from widget output to ANSI escape sequences.
var colorMap = map[string]string{
	"red":     "\033[0;31m",
	"green":   "\033[0;32m",
	"yellow":  "\033[0;33m",
	"blue":    "\033[0;34m",
	"magenta": "\033[0;35m",
	"cyan":    "\033[0;36m",
	"white":   "\033[0;37m",
	"dim":     "\033[2m",
	"gray":    "\033[0;90m",
}

const ansiReset = "\033[0m"

// Render takes ordered widget results grouped by line and produces
// the final colored string for stdout. Widgets on the same line are
// joined with spaces. Lines are separated by \n. Empty lines (all
// widgets nil) are skipped. No trailing newline on the last line.
func Render(lines [][]WidgetResult) string {
	var rendered []string

	for _, line := range lines {
		var parts []string
		for _, wr := range line {
			if wr.Output == nil {
				continue
			}
			text := wr.Output.Text
			if code, ok := colorMap[wr.Output.Color]; ok {
				text = code + text + ansiReset
			}
			parts = append(parts, text)
		}
		if len(parts) == 0 {
			continue
		}
		rendered = append(rendered, strings.Join(parts, " "))
	}

	return strings.Join(rendered, "\n")
}
