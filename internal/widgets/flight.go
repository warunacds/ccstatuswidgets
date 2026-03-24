package widgets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// FlightWidget displays real-time flight status using the AviationStack API.
type FlightWidget struct{}

type aviationStackResponse struct {
	Data []aviationFlight `json:"data"`
}

type aviationFlight struct {
	FlightStatus string         `json:"flight_status"`
	Departure    aviationAirport `json:"departure"`
	Arrival      aviationAirport `json:"arrival"`
}

type aviationAirport struct {
	IATA     string `json:"iata"`
	Actual   string `json:"actual"`
	Estimated string `json:"estimated"`
	Scheduled string `json:"scheduled"`
	Terminal string `json:"terminal"`
	Gate     string `json:"gate"`
}

func (w *FlightWidget) Name() string {
	return "flight"
}

func (w *FlightWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	apiKey, _ := cfg["api_key"].(string)
	flight, _ := cfg["flight"].(string)
	if apiKey == "" || flight == "" {
		return nil, nil
	}

	baseURL := "http://api.aviationstack.com"
	if override, ok := cfg["base_url"].(string); ok && override != "" {
		baseURL = override
	}

	url := fmt.Sprintf("%s/v1/flights?access_key=%s&flight_iata=%s", baseURL, apiKey, flight)

	client := httpclient.New()
	body, err := client.Get(url)
	if err != nil {
		return nil, nil
	}

	var resp aviationStackResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	f := resp.Data[0]
	dep := f.Departure
	arr := f.Arrival

	// Build rich output: ✈ SQ478 SIN ━━━━✈━━━━━ JNB 14:00→18:29 T3/A1
	var parts []string
	parts = append(parts, fmt.Sprintf("✈ %s", flight))

	// Times
	depTime := pickTime(dep.Actual, dep.Estimated, dep.Scheduled)
	arrTime := pickTime(arr.Actual, arr.Estimated, arr.Scheduled)

	// Flight progress bar: SIN ━━━━✈━━━━━ JNB
	if dep.IATA != "" && arr.IATA != "" {
		pct := flightProgress(f, cfg)
		bar := buildFlightBar(pct, 10)
		parts = append(parts, fmt.Sprintf("%s %s %s", dep.IATA, bar, arr.IATA))
	}

	// Times: 14:00→18:29
	if depTime != "" && arrTime != "" {
		parts = append(parts, fmt.Sprintf("%s→%s", depTime, arrTime))
	}

	// Terminal/Gate
	tg := formatTerminalGate(dep.Terminal, dep.Gate)
	if tg != "" {
		parts = append(parts, tg)
	}

	// Status indicator for non-active states
	switch f.FlightStatus {
	case "landed":
		parts = append(parts, "✓")
	case "cancelled":
		parts = append(parts, "✗")
	case "scheduled":
		parts = append(parts, "⏳")
	}

	return &protocol.WidgetOutput{
		Text:  strings.Join(parts, " "),
		Color: "cyan",
	}, nil
}

// pickTime returns the best available time, extracting HH:MM from an ISO timestamp.
func pickTime(actual, estimated, scheduled string) string {
	for _, t := range []string{actual, estimated, scheduled} {
		if t == "" {
			continue
		}
		// AviationStack returns ISO 8601: "2026-03-24T14:00:00+00:00"
		// Extract HH:MM from the T portion
		if idx := strings.Index(t, "T"); idx >= 0 {
			rest := t[idx+1:]
			if len(rest) >= 5 {
				return rest[:5] // "14:00"
			}
		}
	}
	return ""
}

// flightProgress calculates the percentage of flight completed (0.0 to 1.0).
// Uses departure and arrival times to compute elapsed fraction.
// Returns 0 for scheduled, 1 for landed, and computed value for active.
func flightProgress(f aviationFlight, cfg map[string]interface{}) float64 {
	// Allow override for testing
	if v, ok := cfg["_now"]; ok {
		if ts, ok := v.(float64); ok {
			return flightProgressAt(f, time.Unix(int64(ts), 0))
		}
	}
	return flightProgressAt(f, time.Now())
}

func flightProgressAt(f aviationFlight, now time.Time) float64 {
	switch f.FlightStatus {
	case "landed":
		return 1.0
	case "scheduled":
		return 0.0
	case "cancelled":
		return 0.0
	}

	// Parse departure and arrival times
	depStr := pickTimeRaw(f.Departure.Actual, f.Departure.Estimated, f.Departure.Scheduled)
	arrStr := pickTimeRaw(f.Arrival.Actual, f.Arrival.Estimated, f.Arrival.Scheduled)
	if depStr == "" || arrStr == "" {
		return 0.5 // can't compute, show midway
	}

	depT := parseISO(depStr)
	arrT := parseISO(arrStr)
	if depT.IsZero() || arrT.IsZero() || !arrT.After(depT) {
		return 0.5
	}

	total := arrT.Sub(depT).Seconds()
	elapsed := now.Sub(depT).Seconds()
	if elapsed <= 0 {
		return 0.0
	}
	if elapsed >= total {
		return 1.0
	}
	return elapsed / total
}

// pickTimeRaw returns the best available raw ISO timestamp string.
func pickTimeRaw(actual, estimated, scheduled string) string {
	for _, t := range []string{actual, estimated, scheduled} {
		if t != "" {
			return t
		}
	}
	return ""
}

// parseISO parses an ISO 8601 timestamp from AviationStack.
func parseISO(s string) time.Time {
	// Try common formats
	for _, layout := range []string{
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05+00:00",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// buildFlightBar creates a progress bar like ━━━━✈━━━━━ with the plane
// positioned according to the percentage (0.0 to 1.0).
func buildFlightBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	pos := int(pct * float64(width))
	if pos > width {
		pos = width
	}

	var b strings.Builder
	for i := 0; i < pos; i++ {
		b.WriteString("━")
	}
	if pos < width {
		b.WriteString("✈")
		for i := pos + 1; i < width; i++ {
			b.WriteString("─")
		}
	} else {
		// Plane at the end (landed)
		b.WriteString("✈")
	}
	return b.String()
}

// formatTerminalGate builds "T3/A1" from terminal and gate values.
func formatTerminalGate(terminal, gate string) string {
	if terminal == "" && gate == "" {
		return ""
	}
	if terminal != "" && gate != "" {
		return fmt.Sprintf("T%s/%s", terminal, gate)
	}
	if terminal != "" {
		return fmt.Sprintf("T%s", terminal)
	}
	return gate
}
