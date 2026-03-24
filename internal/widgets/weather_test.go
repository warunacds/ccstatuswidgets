package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestWeatherWidget_Name(t *testing.T) {
	w := &WeatherWidget{}
	if w.Name() != "weather" {
		t.Errorf("expected name %q, got %q", "weather", w.Name())
	}
}

func TestWeatherWidget_ReturnsWeatherTextOnSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "  ☀️ +28°C  ")
	}))
	defer srv.Close()

	w := &WeatherWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"city":     "London",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "☀️ +28°C" {
		t.Errorf("expected text %q, got %q", "☀️ +28°C", out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}

func TestWeatherWidget_ReturnsNilOnHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	w := &WeatherWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
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

func TestWeatherWidget_ReadsCityFromConfig(t *testing.T) {
	var requestedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		fmt.Fprint(w, "🌧 +12°C")
	}))
	defer srv.Close()

	w := &WeatherWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"city":     "Tokyo",
	}

	_, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestedPath != "/Tokyo" {
		t.Errorf("expected request path %q, got %q", "/Tokyo", requestedPath)
	}
}

func TestWeatherWidget_UsesImperialUnits(t *testing.T) {
	var requestedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedQuery = r.URL.RawQuery
		fmt.Fprint(w, "☀️ +82°F")
	}))
	defer srv.Close()

	w := &WeatherWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"city":     "NYC",
		"units":    "imperial",
	}

	_, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "format=%c+%t&u"
	if requestedQuery != expected {
		t.Errorf("expected query %q, got %q", expected, requestedQuery)
	}
}

func TestWeatherWidget_FallsBackToEmptyCity(t *testing.T) {
	var requestedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		fmt.Fprint(w, "☀️ +25°C")
	}))
	defer srv.Close()

	w := &WeatherWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
	}

	_, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With no city, path should be just "/"
	if requestedPath != "/" {
		t.Errorf("expected request path %q, got %q", "/", requestedPath)
	}
}
