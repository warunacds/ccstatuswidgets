package widgets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// FlightWidget displays real-time flight status.
// Supports two providers: "aviationstack" (default) and "aerodatabox".
type FlightWidget struct{}

// Unified flight data — both providers parse into this.
type flightData struct {
	Status    string // active, landed, scheduled, cancelled
	DepIATA   string
	ArrIATA   string
	DepTime   string // ISO 8601
	ArrTime   string // ISO 8601
	Terminal  string
	Gate      string
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

	provider, _ := cfg["provider"].(string)
	if provider == "" {
		provider = "aviationstack"
	}

	var fd *flightData
	var err error

	switch provider {
	case "aerodatabox":
		fd, err = fetchAeroDataBox(apiKey, flight, cfg)
	default:
		fd, err = fetchAviationStack(apiKey, flight, cfg)
	}

	if err != nil || fd == nil {
		return nil, nil
	}

	return renderFlight(flight, fd, cfg), nil
}

// --- AviationStack provider ---

type aviationStackResponse struct {
	Data []aviationStackFlight `json:"data"`
}

type aviationStackFlight struct {
	FlightStatus string                `json:"flight_status"`
	Departure    aviationStackAirport  `json:"departure"`
	Arrival      aviationStackAirport  `json:"arrival"`
}

type aviationStackAirport struct {
	IATA      string `json:"iata"`
	Actual    string `json:"actual"`
	Estimated string `json:"estimated"`
	Scheduled string `json:"scheduled"`
	Terminal  string `json:"terminal"`
	Gate      string `json:"gate"`
}

func fetchAviationStack(apiKey, flight string, cfg map[string]interface{}) (*flightData, error) {
	baseURL := "http://api.aviationstack.com"
	if override, ok := cfg["base_url"].(string); ok && override != "" {
		baseURL = override
	}

	url := fmt.Sprintf("%s/v1/flights?access_key=%s&flight_iata=%s", baseURL, apiKey, flight)

	client := httpclient.New()
	body, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	var resp aviationStackResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	f := resp.Data[0]
	return &flightData{
		Status:   f.FlightStatus,
		DepIATA:  f.Departure.IATA,
		ArrIATA:  f.Arrival.IATA,
		DepTime:  pickTimeRaw(f.Departure.Actual, f.Departure.Estimated, f.Departure.Scheduled),
		ArrTime:  pickTimeRaw(f.Arrival.Actual, f.Arrival.Estimated, f.Arrival.Scheduled),
		Terminal: f.Departure.Terminal,
		Gate:     f.Departure.Gate,
	}, nil
}

// --- AeroDataBox provider ---

type aeroDataBoxFlight struct {
	Status    string              `json:"status"`
	Departure aeroDataBoxAirport  `json:"departure"`
	Arrival   aeroDataBoxAirport  `json:"arrival"`
}

type aeroDataBoxAirport struct {
	Airport struct {
		IATA string `json:"iata"`
	} `json:"airport"`
	ActualTime    string `json:"actualTime"`
	EstimatedTime string `json:"estimatedTime"`
	ScheduledTime string `json:"scheduledTime"`
	Terminal      string `json:"terminal"`
	Gate          string `json:"gate"`
}

func fetchAeroDataBox(apiKey, flight string, cfg map[string]interface{}) (*flightData, error) {
	baseURL := "https://aerodatabox.p.rapidapi.com"
	if override, ok := cfg["base_url"].(string); ok && override != "" {
		baseURL = override
	}

	// AeroDataBox requires date-based lookup
	today := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("%s/flights/number/%s/%s", baseURL, flight, today)

	client := httpclient.New()

	// AeroDataBox uses RapidAPI — need custom headers, so build request manually
	req, err := newAeroRequest(url, apiKey)
	if err != nil {
		return nil, err
	}

	body, err := doRequest(client, req)
	if err != nil {
		return nil, err
	}

	var flights []aeroDataBoxFlight
	if err := json.Unmarshal(body, &flights); err != nil {
		return nil, err
	}

	if len(flights) == 0 {
		return nil, nil
	}

	f := flights[0]

	// Map AeroDataBox status to our standard statuses
	status := mapAeroStatus(f.Status)

	return &flightData{
		Status:   status,
		DepIATA:  f.Departure.Airport.IATA,
		ArrIATA:  f.Arrival.Airport.IATA,
		DepTime:  pickTimeRaw(f.Departure.ActualTime, f.Departure.EstimatedTime, f.Departure.ScheduledTime),
		ArrTime:  pickTimeRaw(f.Arrival.ActualTime, f.Arrival.EstimatedTime, f.Arrival.ScheduledTime),
		Terminal: f.Departure.Terminal,
		Gate:     f.Departure.Gate,
	}, nil
}

func mapAeroStatus(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "airborne"), strings.Contains(s, "en route"), strings.Contains(s, "active"):
		return "active"
	case strings.Contains(s, "landed"), strings.Contains(s, "arrived"):
		return "landed"
	case strings.Contains(s, "cancelled"), strings.Contains(s, "canceled"):
		return "cancelled"
	case strings.Contains(s, "scheduled"), strings.Contains(s, "expected"):
		return "scheduled"
	default:
		return s
	}
}

// newAeroRequest builds an HTTP request with RapidAPI headers for AeroDataBox.
func newAeroRequest(url, apiKey string) (*httpRequest, error) {
	return &httpRequest{
		URL: url,
		Headers: map[string]string{
			"X-RapidAPI-Key":  apiKey,
			"X-RapidAPI-Host": "aerodatabox.p.rapidapi.com",
		},
	}, nil
}

// httpRequest holds a URL and custom headers for requests that need more than a simple GET.
type httpRequest struct {
	URL     string
	Headers map[string]string
}

// doRequest executes an httpRequest using the httpclient's underlying HTTP client.
func doRequest(client *httpclient.Client, req *httpRequest) ([]byte, error) {
	return client.GetWithHeaders(req.URL, req.Headers)
}

// --- Shared rendering ---

func renderFlight(flight string, fd *flightData, cfg map[string]interface{}) *protocol.WidgetOutput {
	var parts []string
	parts = append(parts, fmt.Sprintf("✈ %s", flight))

	// Times
	depTime := pickTimeFromISO(fd.DepTime)
	arrTime := pickTimeFromISO(fd.ArrTime)

	// Flight progress bar: SIN ━━━━✈━━━━━ JNB
	if fd.DepIATA != "" && fd.ArrIATA != "" {
		pct := computeProgress(fd, cfg)
		bar := buildFlightBar(pct, 10)
		parts = append(parts, fmt.Sprintf("%s %s %s", fd.DepIATA, bar, fd.ArrIATA))
	}

	// Times: 14:00→18:29
	if depTime != "" && arrTime != "" {
		parts = append(parts, fmt.Sprintf("%s→%s", depTime, arrTime))
	}

	// Terminal/Gate
	tg := formatTerminalGate(fd.Terminal, fd.Gate)
	if tg != "" {
		parts = append(parts, tg)
	}

	// Status indicator
	switch fd.Status {
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
	}
}

func computeProgress(fd *flightData, cfg map[string]interface{}) float64 {
	f := aviationStackFlight{
		FlightStatus: fd.Status,
		Departure: aviationStackAirport{
			Actual:    fd.DepTime,
			Estimated: fd.DepTime,
		},
		Arrival: aviationStackAirport{
			Actual:    fd.ArrTime,
			Estimated: fd.ArrTime,
		},
	}
	return flightProgress(f, cfg)
}

// pickTimeFromISO extracts HH:MM from an ISO timestamp string.
func pickTimeFromISO(s string) string {
	if s == "" {
		return ""
	}
	if idx := strings.Index(s, "T"); idx >= 0 {
		rest := s[idx+1:]
		if len(rest) >= 5 {
			return rest[:5]
		}
	}
	return ""
}

// pickTime returns the best available time, extracting HH:MM from an ISO timestamp.
func pickTime(actual, estimated, scheduled string) string {
	raw := pickTimeRaw(actual, estimated, scheduled)
	return pickTimeFromISO(raw)
}

// flightProgress calculates the percentage of flight completed (0.0 to 1.0).
func flightProgress(f aviationStackFlight, cfg map[string]interface{}) float64 {
	if v, ok := cfg["_now"]; ok {
		if ts, ok := v.(float64); ok {
			return flightProgressAt(f, time.Unix(int64(ts), 0))
		}
	}
	return flightProgressAt(f, time.Now())
}

func flightProgressAt(f aviationStackFlight, now time.Time) float64 {
	switch f.FlightStatus {
	case "landed":
		return 1.0
	case "scheduled", "cancelled":
		return 0.0
	}

	depStr := pickTimeRaw(f.Departure.Actual, f.Departure.Estimated, f.Departure.Scheduled)
	arrStr := pickTimeRaw(f.Arrival.Actual, f.Arrival.Estimated, f.Arrival.Scheduled)
	if depStr == "" || arrStr == "" {
		return 0.5
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

// parseISO parses an ISO 8601 timestamp.
func parseISO(s string) time.Time {
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

// buildFlightBar creates a progress bar like ━━━━✈━━━━━
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
