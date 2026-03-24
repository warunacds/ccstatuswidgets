package widgets

import (
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestNowPlayingWidget_Name(t *testing.T) {
	w := &NowPlayingWidget{}
	if w.Name() != "now-playing" {
		t.Errorf("expected name %q, got %q", "now-playing", w.Name())
	}
}

func TestNowPlayingWidget_ParsesCommandOutput(t *testing.T) {
	w := &NowPlayingWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"command": "echo",
		"args":    "Artist - Song",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "\u266a Artist - Song" {
		t.Errorf("expected text %q, got %q", "\u266a Artist - Song", out.Text)
	}
	if out.Color != "magenta" {
		t.Errorf("expected color %q, got %q", "magenta", out.Color)
	}
}

func TestNowPlayingWidget_ReturnsNilWhenCommandFails(t *testing.T) {
	w := &NowPlayingWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"command": "false",
		"args":    "",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when command fails, got %+v", out)
	}
}

func TestNowPlayingWidget_TruncatesLongText(t *testing.T) {
	w := &NowPlayingWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"command": "echo",
		"args":    "Some Really Long Artist Name - Some Really Long Song Title That Keeps Going",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	// The text after the "♪ " prefix should be ~30 chars, plus the "..." suffix
	// Total text = "♪ " + truncated + "..."
	prefix := "\u266a "
	if !strings.HasPrefix(out.Text, prefix) {
		t.Errorf("expected text to start with %q, got %q", prefix, out.Text)
	}

	body := strings.TrimPrefix(out.Text, prefix)
	if !strings.HasSuffix(body, "...") {
		t.Errorf("expected truncated text to end with '...', got %q", body)
	}

	// The body (without "...") should be ~30 chars
	bodyWithoutEllipsis := strings.TrimSuffix(body, "...")
	if len(bodyWithoutEllipsis) > 30 {
		t.Errorf("expected body to be at most 30 chars, got %d: %q", len(bodyWithoutEllipsis), bodyWithoutEllipsis)
	}
}

func TestNowPlayingWidget_OutputHasMusicPrefixAndMagentaColor(t *testing.T) {
	w := &NowPlayingWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"command": "echo",
		"args":    "Radiohead - Creep",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if !strings.HasPrefix(out.Text, "\u266a ") {
		t.Errorf("expected text to start with music note prefix, got %q", out.Text)
	}
	if out.Color != "magenta" {
		t.Errorf("expected color %q, got %q", "magenta", out.Color)
	}
}

func TestNowPlayingWidget_ReturnsNilWhenOutputEmpty(t *testing.T) {
	w := &NowPlayingWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"command": "echo",
		"args":    "",
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output for empty text, got %+v", out)
	}
}
