package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	// Should contain: ✈ SQ478 SIN→JNB 14:00→18:29 T3/A1 ⬆
	if !strings.Contains(out.Text, "SIN→JNB") {
		t.Errorf("expected route in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "14:00→18:29") {
		t.Errorf("expected times in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "T3/A1") {
		t.Errorf("expected terminal/gate in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "⬆") {
		t.Errorf("expected active indicator in output, got %q", out.Text)
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
