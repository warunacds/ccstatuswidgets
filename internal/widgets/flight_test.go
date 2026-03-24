package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

const richFlightJSON = `{
	"data": [{
		"flight_status": "active",
		"departure": {
			"iata": "SIN",
			"actual": "2026-03-24T14:00:00+00:00",
			"terminal": "3",
			"gate": "A1"
		},
		"arrival": {
			"iata": "JNB",
			"estimated": "2026-03-24T18:29:00+00:00",
			"terminal": "A"
		}
	}]
}`

func TestFlightWidget_Name(t *testing.T) {
	w := &FlightWidget{}
	if w.Name() != "flight" {
		t.Errorf("expected name %q, got %q", "flight", w.Name())
	}
}

func TestFlightWidget_RichOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, richFlightJSON)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "SQ478",
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	// Should contain: ✈ SQ478 SIN ━━━✈━━━ JNB 14:00→18:29 T3/A1
	if !strings.Contains(out.Text, "SIN") || !strings.Contains(out.Text, "JNB") {
		t.Errorf("expected airports in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "14:00→18:29") {
		t.Errorf("expected times in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "T3/A1") {
		t.Errorf("expected terminal/gate in output, got %q", out.Text)
	}
	if out.Color != "cyan" {
		t.Errorf("expected color %q, got %q", "cyan", out.Color)
	}
}

func TestFlightWidget_LandedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"flight_status":"landed","departure":{"iata":"SIN"},"arrival":{"iata":"JNB"}}]}`)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "UL504",
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.Text, "✓") {
		t.Errorf("expected landed checkmark, got %q", out.Text)
	}
}

func TestFlightWidget_CancelledStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"flight_status":"cancelled","departure":{"iata":"SIN"},"arrival":{"iata":"JNB"}}]}`)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "UL504",
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.Text, "✗") {
		t.Errorf("expected cancelled cross, got %q", out.Text)
	}
}

func TestFlightWidget_MinimalData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"flight_status":"scheduled"}]}`)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "UL504",
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.Contains(out.Text, "✈ UL504") {
		t.Errorf("expected flight number, got %q", out.Text)
	}
}

func TestFlightWidget_ReturnsNilWhenNoApiKey(t *testing.T) {
	w := &FlightWidget{}
	cfg := map[string]interface{}{"flight": "UL504"}
	out, _ := w.Render(&protocol.StatusLineInput{}, cfg)
	if out != nil {
		t.Errorf("expected nil, got %+v", out)
	}
}

func TestFlightWidget_ReturnsNilWhenNoFlight(t *testing.T) {
	w := &FlightWidget{}
	cfg := map[string]interface{}{"api_key": "test_key"}
	out, _ := w.Render(&protocol.StatusLineInput{}, cfg)
	if out != nil {
		t.Errorf("expected nil, got %+v", out)
	}
}

func TestFlightWidget_ReturnsNilOnHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "UL504",
		"base_url": srv.URL,
	}
	out, _ := w.Render(&protocol.StatusLineInput{}, cfg)
	if out != nil {
		t.Errorf("expected nil, got %+v", out)
	}
}

func TestBuildFlightBar(t *testing.T) {
	// 0% — plane at start
	bar := buildFlightBar(0, 10)
	if !strings.HasPrefix(bar, "✈") {
		t.Errorf("0%% bar should start with plane, got %q", bar)
	}

	// 50% — plane in middle
	bar = buildFlightBar(0.5, 10)
	if !strings.Contains(bar, "━") || !strings.Contains(bar, "✈") {
		t.Errorf("50%% bar missing expected chars, got %q", bar)
	}
	// Count rune position of plane
	runePos := 0
	for _, r := range bar {
		if string(r) == "✈" {
			break
		}
		runePos++
	}
	if runePos < 3 || runePos > 7 {
		t.Errorf("50%% bar plane should be near middle, got at rune pos %d in %q", runePos, bar)
	}

	// 100% — plane at end
	bar = buildFlightBar(1.0, 10)
	if !strings.HasSuffix(bar, "✈") {
		t.Errorf("100%% bar should end with plane, got %q", bar)
	}
}

func TestFlightProgress_Statuses(t *testing.T) {
	landed := aviationFlight{FlightStatus: "landed"}
	if p := flightProgressAt(landed, time.Now()); p != 1.0 {
		t.Errorf("landed should be 1.0, got %f", p)
	}

	scheduled := aviationFlight{FlightStatus: "scheduled"}
	if p := flightProgressAt(scheduled, time.Now()); p != 0.0 {
		t.Errorf("scheduled should be 0.0, got %f", p)
	}
}

func TestFlightProgress_Active(t *testing.T) {
	f := aviationFlight{
		FlightStatus: "active",
		Departure: aviationAirport{
			Actual: "2026-03-24T10:00:00+00:00",
		},
		Arrival: aviationAirport{
			Estimated: "2026-03-24T20:00:00+00:00",
		},
	}

	// 5 hours into a 10-hour flight = 50%
	now := time.Date(2026, 3, 24, 15, 0, 0, 0, time.UTC)
	pct := flightProgressAt(f, now)
	if pct < 0.45 || pct > 0.55 {
		t.Errorf("expected ~50%%, got %f", pct)
	}

	// At departure = 0%
	pct = flightProgressAt(f, time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC))
	if pct != 0.0 {
		t.Errorf("expected 0%%, got %f", pct)
	}

	// At arrival = 100%
	pct = flightProgressAt(f, time.Date(2026, 3, 24, 20, 0, 0, 0, time.UTC))
	if pct != 1.0 {
		t.Errorf("expected 100%%, got %f", pct)
	}
}

func TestFlightWidget_RichOutputWithProgressBar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, richFlightJSON)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "SQ478",
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain the progress bar with SIN ... JNB
	if !strings.Contains(out.Text, "SIN") || !strings.Contains(out.Text, "JNB") {
		t.Errorf("expected airports in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "✈") {
		t.Errorf("expected plane in progress bar, got %q", out.Text)
	}
}

func TestPickTime(t *testing.T) {
	if got := pickTime("2026-03-24T14:00:00+00:00", "", ""); got != "14:00" {
		t.Errorf("expected 14:00, got %q", got)
	}
	if got := pickTime("", "2026-03-24T18:29:00+00:00", ""); got != "18:29" {
		t.Errorf("expected 18:29, got %q", got)
	}
	if got := pickTime("", "", ""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFormatTerminalGate(t *testing.T) {
	if got := formatTerminalGate("3", "A1"); got != "T3/A1" {
		t.Errorf("expected T3/A1, got %q", got)
	}
	if got := formatTerminalGate("A", ""); got != "TA" {
		t.Errorf("expected TA, got %q", got)
	}
	if got := formatTerminalGate("", ""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
