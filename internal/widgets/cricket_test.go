package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

const espnLiveJSON = `{
	"events": [{
		"name": "Sri Lanka vs Australia",
		"status": {"type": {"id": "2", "description": "In Progress", "detail": "42.1 ov", "state": "in"}},
		"competitions": [{
			"competitors": [
				{"team": {"abbreviation": "SL", "displayName": "Sri Lanka"}, "score": "245/3"},
				{"team": {"abbreviation": "AUS", "displayName": "Australia"}, "score": ""}
			],
			"status": {"type": {"state": "in"}}
		}]
	}]
}`

const espnCompletedJSON = `{
	"events": [{
		"name": "India vs England",
		"status": {"type": {"id": "3", "description": "India won by 5 wickets", "detail": "Result", "state": "post"}},
		"competitions": [{
			"competitors": [
				{"team": {"abbreviation": "IND", "displayName": "India"}, "score": "280/5"},
				{"team": {"abbreviation": "ENG", "displayName": "England"}, "score": "275"}
			],
			"status": {"type": {"state": "post"}}
		}]
	}]
}`

const espnScheduledJSON = `{
	"events": [{
		"name": "RCB vs SRH",
		"status": {"type": {"id": "1", "description": "Scheduled", "detail": "7:30 PM", "state": "pre"}},
		"competitions": [{
			"competitors": [
				{"team": {"abbreviation": "RCB", "displayName": "Royal Challengers Bengaluru"}, "score": ""},
				{"team": {"abbreviation": "SRH", "displayName": "Sunrisers Hyderabad"}, "score": ""}
			],
			"status": {"type": {"state": "pre"}}
		}]
	}]
}`

func TestCricketWidget_Name(t *testing.T) {
	w := &CricketWidget{}
	if w.Name() != "cricket" {
		t.Errorf("expected cricket, got %s", w.Name())
	}
}

func TestCricketWidget_LiveScore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, espnLiveJSON)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"leagues":  []interface{}{"8048"},
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.Contains(out.Text, "SL") {
		t.Errorf("expected SL in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "245/3") {
		t.Errorf("expected score in output, got %q", out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected green, got %s", out.Color)
	}
}

func TestCricketWidget_CompletedMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, espnCompletedJSON)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"leagues":  []interface{}{"8048"},
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.Contains(out.Text, "IND v ENG") {
		t.Errorf("expected teams in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "won") {
		t.Errorf("expected result in output, got %q", out.Text)
	}
}

func TestCricketWidget_ScheduledMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, espnScheduledJSON)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"leagues":  []interface{}{"8048"},
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.Contains(out.Text, "RCB v SRH") {
		t.Errorf("expected teams in output, got %q", out.Text)
	}
}

func TestCricketWidget_TeamFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, espnLiveJSON)
	}))
	defer srv.Close()

	w := &CricketWidget{}

	// Filter for SL — should match
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"leagues":  []interface{}{"8048"},
		"team":     "SL",
	}
	out, _ := w.Render(&protocol.StatusLineInput{}, cfg)
	if out == nil {
		t.Fatal("expected match for SL filter")
	}

	// Filter for IND — should not match
	cfg["team"] = "IND"
	out, _ = w.Render(&protocol.StatusLineInput{}, cfg)
	if out != nil {
		t.Errorf("expected nil for IND filter, got %q", out.Text)
	}
}

func TestCricketWidget_NoMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"events":[]}`)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"leagues":  []interface{}{"8048"},
	}

	out, _ := w.Render(&protocol.StatusLineInput{}, cfg)
	if out != nil {
		t.Errorf("expected nil for empty events, got %+v", out)
	}
}

func TestCricketWidget_HTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	w := &CricketWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
		"leagues":  []interface{}{"8048"},
	}

	out, _ := w.Render(&protocol.StatusLineInput{}, cfg)
	if out != nil {
		t.Errorf("expected nil on HTTP failure, got %+v", out)
	}
}

func TestCricketWidget_NoConfigNeeded(t *testing.T) {
	// Should not panic with nil config
	w := &CricketWidget{}
	_, err := w.Render(&protocol.StatusLineInput{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
