package widgets

import (
	"fmt"
	"math"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// Usage5hWidget displays the 5-hour rate limit usage bar.
type Usage5hWidget struct{}

func (w *Usage5hWidget) Name() string {
	return "usage-5h"
}

func (w *Usage5hWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if input.RateLimits == nil || input.RateLimits.FiveHour == nil {
		return nil, nil
	}
	return renderUsage("5h", 5*time.Hour, input.RateLimits.FiveHour, cfg)
}

// Usage7dWidget displays the 7-day rate limit usage bar.
type Usage7dWidget struct{}

func (w *Usage7dWidget) Name() string {
	return "usage-7d"
}

func (w *Usage7dWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if input.RateLimits == nil || input.RateLimits.SevenDay == nil {
		return nil, nil
	}
	return renderUsage("7d", 7*24*time.Hour, input.RateLimits.SevenDay, cfg)
}

func renderUsage(label string, windowDuration time.Duration, rl *protocol.RateLimit, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	pct := rl.UsedPercentage

	barLength := 10
	showPercentage := true
	showPace := true
	if cfg != nil {
		if bl, ok := cfg["bar_length"].(float64); ok && bl > 0 {
			barLength = int(bl)
		}
		if sp, ok := cfg["show_percentage"].(bool); ok {
			showPercentage = sp
		}
		if sp, ok := cfg["show_pace"].(bool); ok {
			showPace = sp
		}
	}

	bar := BuildBar(pct, barLength)
	color := BarColor(pct)
	resetStr := formatRelativeTime(rl.ResetsAt)

	// Calculate pace: used% - elapsed%
	elapsed := calcElapsedPercentage(rl.ResetsAt, windowDuration)
	pace := int(math.Round(pct)) - int(math.Round(elapsed))

	text := label + " " + bar
	if showPercentage {
		text += fmt.Sprintf(" %d%%", int(math.Round(pct)))
	}
	if showPace {
		if pace >= 0 {
			text += fmt.Sprintf(" +%d%%", pace)
		} else {
			text += fmt.Sprintf(" %d%%", pace)
		}
	}
	text += fmt.Sprintf(" \u21bb%s", resetStr)

	return &protocol.WidgetOutput{
		Text:  text,
		Color: color,
	}, nil
}

func calcElapsedPercentage(resetsAt int64, windowDuration time.Duration) float64 {
	remaining := time.Until(time.Unix(resetsAt, 0))
	if remaining < 0 {
		return 100
	}
	elapsed := windowDuration - remaining
	return float64(elapsed) / float64(windowDuration) * 100
}

func formatRelativeTime(unixTS int64) string {
	remaining := time.Until(time.Unix(unixTS, 0))
	if remaining < 0 {
		return "0m"
	}

	// Round up to the nearest minute to avoid off-by-one from sub-second drift.
	remaining = remaining.Round(time.Minute)

	totalMinutes := int(remaining.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
