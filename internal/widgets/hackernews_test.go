package widgets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestHackernewsWidget_Name(t *testing.T) {
	w := &HackernewsWidget{}
	if w.Name() != "hackernews" {
		t.Errorf("expected name %q, got %q", "hackernews", w.Name())
	}
}

func TestHackernewsWidget_ReturnsTopStoryTitle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v0/topstories.json":
			json.NewEncoder(w).Encode([]int{12345})
		case "/v0/item/12345.json":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"title": "Go 1.26 Released",
				"score": 342,
				"url":   "https://go.dev/blog/go1.26",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	w := &HackernewsWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "HN: Go 1.26 Released" {
		t.Errorf("expected %q, got %q", "HN: Go 1.26 Released", out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}

func TestHackernewsWidget_ShowsScoreWhenEnabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v0/topstories.json":
			json.NewEncoder(w).Encode([]int{99})
		case "/v0/item/99.json":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"title": "Rust is Cool",
				"score": 512,
				"url":   "https://example.com",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	w := &HackernewsWidget{}
	cfg := map[string]interface{}{
		"base_url":   srv.URL,
		"show_score": true,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := "HN: Rust is Cool (512pts)"
	if out.Text != expected {
		t.Errorf("expected %q, got %q", expected, out.Text)
	}
}

func TestHackernewsWidget_TruncatesLongTitles(t *testing.T) {
	longTitle := "This Is an Extremely Long Hacker News Story Title That Should Be Truncated"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v0/topstories.json":
			json.NewEncoder(w).Encode([]int{1})
		case "/v0/item/1.json":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"title": longTitle,
				"score": 10,
				"url":   "https://example.com",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	w := &HackernewsWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// "HN: " is 4 chars, title should be truncated to ~40 chars + "..."
	// Total text should not exceed "HN: " + 40 + "..." = 47 chars
	titlePart := strings.TrimPrefix(out.Text, "HN: ")
	if len(titlePart) > 43 { // 40 + "..."
		t.Errorf("title not truncated: %q (len %d)", titlePart, len(titlePart))
	}
	if !strings.HasSuffix(titlePart, "...") {
		t.Errorf("expected truncated title to end with '...', got %q", titlePart)
	}
}

func TestHackernewsWidget_ReturnsNilOnHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	w := &HackernewsWidget{}
	cfg := map[string]interface{}{
		"base_url": srv.URL,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output on HTTP failure, got %+v", out)
	}
}
