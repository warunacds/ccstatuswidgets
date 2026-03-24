package widgets

import (
	"encoding/json"
	"fmt"
	"strings"

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

	// Build rich output: ✈ SQ478 SIN→JNB 14:00→18:29 T3/A1
	var parts []string
	parts = append(parts, fmt.Sprintf("✈ %s", flight))

	// Route: SIN→JNB
	if dep.IATA != "" && arr.IATA != "" {
		parts = append(parts, fmt.Sprintf("%s→%s", dep.IATA, arr.IATA))
	}

	// Times: departure→arrival (use actual if available, then estimated, then scheduled)
	depTime := pickTime(dep.Actual, dep.Estimated, dep.Scheduled)
	arrTime := pickTime(arr.Actual, arr.Estimated, arr.Scheduled)
	if depTime != "" && arrTime != "" {
		parts = append(parts, fmt.Sprintf("%s→%s", depTime, arrTime))
	}

	// Terminal/Gate
	tg := formatTerminalGate(dep.Terminal, dep.Gate)
	if tg != "" {
		parts = append(parts, tg)
	}

	// Status indicator
	switch f.FlightStatus {
	case "active":
		parts = append(parts, "⬆")
	case "landed":
		parts = append(parts, "✓")
	case "cancelled":
		parts = append(parts, "✗")
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
