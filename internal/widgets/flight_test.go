package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestFlightWidget_Name(t *testing.T) {
	w := &FlightWidget{}
	if w.Name() != "flight" {
		t.Errorf("expected name %q, got %q", "flight", w.Name())
	}
}

func TestFlightWidget_ReturnsFlightStatusOnSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"flight_status":"active","flight":{"iata":"UL504"}}]}`)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "UL504",
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "✈ UL504 ⬆ active" {
		t.Errorf("expected text %q, got %q", "✈ UL504 ⬆ active", out.Text)
	}
	if out.Color != "cyan" {
		t.Errorf("expected color %q, got %q", "cyan", out.Color)
	}
}

func TestFlightWidget_ReturnsNilWhenNoApiKey(t *testing.T) {
	w := &FlightWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"flight": "UL504",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when no api_key, got %+v", out)
	}
}

func TestFlightWidget_ReturnsNilWhenNoFlight(t *testing.T) {
	w := &FlightWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key": "test_key",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when no flight, got %+v", out)
	}
}

func TestFlightWidget_ReturnsNilOnHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	w := &FlightWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test_key",
		"flight":   "UL504",
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output on HTTP failure, got %+v", out)
	}
}

func TestFlightWidget_ParsesDifferentStatuses(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"active", "✈ UL504 ⬆ active"},
		{"landed", "✈ UL504 landed"},
		{"scheduled", "✈ UL504 scheduled"},
		{"cancelled", "✈ UL504 cancelled"},
	}

	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"data":[{"flight_status":"%s","flight":{"iata":"UL504"}}]}`, tc.status)
			}))
			defer srv.Close()

			w := &FlightWidget{}
			input := &protocol.StatusLineInput{}
			cfg := map[string]interface{}{
				"api_key":  "test_key",
				"flight":   "UL504",
				"base_url": srv.URL,
			}

			out, err := w.Render(input, cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.Text != tc.expected {
				t.Errorf("expected text %q, got %q", tc.expected, out.Text)
			}
		})
	}
}
