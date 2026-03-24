package widgets

import "math"

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
	bar := ""
	for i := 0; i < filled; i++ {
		bar += filledBlock
	}
	for i := filled; i < width; i++ {
		bar += emptyBlock
	}
	return bar
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
