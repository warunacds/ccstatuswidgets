package widgets

import (
	"math"
	"strings"
)

const (
	filledBlock = "\u2588"
	emptyBlock  = "\u2591"
)

// BuildBar creates a bar string of the given width filled proportionally to pct (0-100).
func BuildBar(pct float64, width int) string {
	filled := int(math.Round(pct / 100 * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat(filledBlock, filled) + strings.Repeat(emptyBlock, width-filled)
}

// BarColor returns a color name based on the percentage threshold.
func BarColor(pct float64) string {
	switch {
	case pct >= 80:
		return "red"
	case pct >= 50:
		return "yellow"
	default:
		return "green"
	}
}
