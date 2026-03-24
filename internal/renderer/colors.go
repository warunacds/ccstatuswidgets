package renderer

import (
	"fmt"
	"strconv"
	"strings"
)

// namedFgCodes maps color names to their ANSI foreground code fragments.
var namedFgCodes = map[string]string{
	"red":     "31",
	"green":   "32",
	"yellow":  "33",
	"blue":    "34",
	"magenta": "35",
	"cyan":    "36",
	"white":   "37",
	"dim":     "2",
	"gray":    "90",
}

// namedBgCodes maps color names to their ANSI background code fragments.
// "dim" is a style modifier, not a color, so it has no background equivalent.
var namedBgCodes = map[string]string{
	"red":     "41",
	"green":   "42",
	"yellow":  "43",
	"blue":    "44",
	"magenta": "45",
	"cyan":    "46",
	"white":   "47",
	"dim":     "",
	"gray":    "100",
}

// FgCode converts a color string to an ANSI foreground code fragment.
// Supported formats: named colors, 256-color ("196" or "color:42"),
// and hex truecolor ("#RRGGBB"). Returns "" for empty or invalid input.
func FgCode(color string) string {
	return colorCode(color, true)
}

// BgCode converts a color string to an ANSI background code fragment.
// Supported formats: named colors, 256-color ("196" or "color:42"),
// and hex truecolor ("#RRGGBB"). Returns "" for empty or invalid input.
func BgCode(color string) string {
	return colorCode(color, false)
}

func colorCode(color string, fg bool) string {
	if color == "" {
		return ""
	}

	// Named colors.
	if fg {
		if code, ok := namedFgCodes[color]; ok {
			return code
		}
	} else {
		if code, ok := namedBgCodes[color]; ok {
			return code
		}
	}

	// Hex truecolor: #RRGGBB
	if strings.HasPrefix(color, "#") && len(color) == 7 {
		r, g, b, err := parseHex(color)
		if err != nil {
			return ""
		}
		if fg {
			return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
		}
		return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	}

	// 256-color: "color:N" prefix format.
	if strings.HasPrefix(color, "color:") {
		numStr := color[6:]
		n, err := strconv.Atoi(numStr)
		if err != nil || n < 0 || n > 255 {
			return ""
		}
		if fg {
			return fmt.Sprintf("38;5;%d", n)
		}
		return fmt.Sprintf("48;5;%d", n)
	}

	// 256-color: plain numeric string.
	n, err := strconv.Atoi(color)
	if err != nil || n < 0 || n > 255 {
		return ""
	}
	if fg {
		return fmt.Sprintf("38;5;%d", n)
	}
	return fmt.Sprintf("48;5;%d", n)
}

// parseHex parses a "#RRGGBB" string into r, g, b components.
func parseHex(hex string) (r, g, b uint8, err error) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}
	rv, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	gv, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	bv, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	return uint8(rv), uint8(gv), uint8(bv), nil
}
