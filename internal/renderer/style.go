package renderer

import "strings"

// WidgetStyle combines foreground, background, and text formatting
// into a single ANSI escape sequence via Prefix().
type WidgetStyle struct {
	Fg        string // color string: named ("red"), 256-color ("196"), or hex ("#ff6b6b")
	Bg        string // color string: same formats as Fg
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
}

// AnsiReset is the ANSI escape sequence to reset all styling.
const AnsiReset = "\033[0m"

// HasStyle returns true if any style property is set.
func (s WidgetStyle) HasStyle() bool {
	return s.Fg != "" || s.Bg != "" || s.Bold || s.Dim || s.Italic || s.Underline
}

// Prefix returns the full ANSI escape sequence prefix for this style.
// Returns "" if no style is set.
func (s WidgetStyle) Prefix() string {
	if !s.HasStyle() {
		return ""
	}

	var parts []string

	if s.Bold {
		parts = append(parts, "1")
	}
	if s.Dim {
		parts = append(parts, "2")
	}
	if s.Italic {
		parts = append(parts, "3")
	}
	if s.Underline {
		parts = append(parts, "4")
	}
	if fg := FgCode(s.Fg); fg != "" {
		parts = append(parts, fg)
	}
	if bg := BgCode(s.Bg); bg != "" {
		parts = append(parts, bg)
	}

	if len(parts) == 0 {
		return ""
	}
	return "\033[" + strings.Join(parts, ";") + "m"
}
