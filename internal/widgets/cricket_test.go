package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestCricketWidget_Name(t *testing.T) {
	w := &CricketWidget{}
	if w.Name() != "cricket" {
		t.Errorf("expected name %q, got %q", "cricket", w.Name())
	}
}

func TestCricketWidget_ReturnsLiveScore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"data": [{
				"name": "Sri Lanka vs Australia",
				"status": "Sri Lanka 245/3 (42.1 ov)",
				"matchType": "odi",
				"teams": ["Sri Lanka", "Australia"],
				"score": [
					{"r": 245, "w": 3, "o": 42.1, "inning": "Sri Lanka Inning 1"}
				],
				"matchStarted": true,
				"matchEnded": false
			}]
		}`)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test-key",
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "\U0001F3CF SL 245/3 (42.1)"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestCricketWidget_FiltersByTeam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"data": [
				{
					"name": "India vs England",
					"status": "India 180/4 (35.0 ov)",
					"matchType": "odi",
					"teams": ["India", "England"],
					"score": [
						{"r": 180, "w": 4, "o": 35.0, "inning": "India Inning 1"}
					],
					"matchStarted": true,
					"matchEnded": false
				},
				{
					"name": "Sri Lanka vs Australia",
					"status": "Sri Lanka 245/3 (42.1 ov)",
					"matchType": "odi",
					"teams": ["Sri Lanka", "Australia"],
					"score": [
						{"r": 245, "w": 3, "o": 42.1, "inning": "Sri Lanka Inning 1"}
					],
					"matchStarted": true,
					"matchEnded": false
				}
			]
		}`)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test-key",
		"base_url": srv.URL,
		"team":     "SL",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "\U0001F3CF SL 245/3 (42.1)"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}

func TestCricketWidget_ReturnsNilWhenNoAPIKey(t *testing.T) {
	w := &CricketWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when no api_key, got %+v", out)
	}
}

func TestCricketWidget_ReturnsNilWhenNoMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data": []}`)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test-key",
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when no matches, got %+v", out)
	}
}

func TestCricketWidget_ReturnsNilOnHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test-key",
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

func TestCricketWidget_HandlesCompletedMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"data": [{
				"name": "Sri Lanka vs Australia",
				"status": "Sri Lanka won by 5 wickets",
				"matchType": "odi",
				"teams": ["Sri Lanka", "Australia"],
				"score": [
					{"r": 280, "w": 10, "o": 48.3, "inning": "Australia Inning 1"},
					{"r": 281, "w": 5, "o": 45.2, "inning": "Sri Lanka Inning 1"}
				],
				"matchStarted": true,
				"matchEnded": true
			}]
		}`)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"api_key":  "test-key",
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "\U0001F3CF SL v AUS - Sri Lanka won by 5 wickets"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}
