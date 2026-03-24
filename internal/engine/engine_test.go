package engine

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/cache"
	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
	"github.com/warunacds/ccstatuswidgets/internal/renderer"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

// --- Mock widgets ---

// fastWidget returns immediately with a fixed output.
type fastWidget struct {
	name   string
	output *protocol.WidgetOutput
}

func (w *fastWidget) Name() string { return w.name }
func (w *fastWidget) Render(_ *protocol.StatusLineInput, _ map[string]interface{}) (*protocol.WidgetOutput, error) {
	return w.output, nil
}

// slowWidget sleeps longer than any reasonable timeout.
type slowWidget struct {
	name  string
	delay time.Duration
}

func (w *slowWidget) Name() string { return w.name }
func (w *slowWidget) Render(_ *protocol.StatusLineInput, _ map[string]interface{}) (*protocol.WidgetOutput, error) {
	time.Sleep(w.delay)
	return &protocol.WidgetOutput{Text: "slow-done", Color: "green"}, nil
}

// errorWidget always returns an error.
type errorWidget struct {
	name string
}

func (w *errorWidget) Name() string { return w.name }
func (w *errorWidget) Render(_ *protocol.StatusLineInput, _ map[string]interface{}) (*protocol.WidgetOutput, error) {
	return nil, errors.New("widget error")
}

// --- Helpers ---

func testInput() *protocol.StatusLineInput {
	return &protocol.StatusLineInput{
		Model:     protocol.ModelInfo{ID: "test-model", DisplayName: "Test"},
		SessionID: "test-session",
	}
}

func testConfig(widgets ...string) *config.Config {
	return &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: widgets},
		},
		Widgets: map[string]map[string]interface{}{},
	}
}

func multiLineConfig(lines ...[]string) *config.Config {
	lcs := make([]config.LineConfig, len(lines))
	for i, ws := range lines {
		lcs[i] = config.LineConfig{Widgets: ws}
	}
	return &config.Config{
		TimeoutMs: 500,
		Lines:     lcs,
		Widgets:   map[string]map[string]interface{}{},
	}
}

func seedCache(c *cache.Cache, key string, out *protocol.WidgetOutput) {
	data, _ := json.Marshal(out)
	c.Set(key, data, 5*time.Minute)
}

// --- Tests ---

func TestRunConcurrentCollectsResults(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "alpha", output: &protocol.WidgetOutput{Text: "A", Color: "green"}})
	reg.Register(&fastWidget{name: "beta", output: &protocol.WidgetOutput{Text: "B", Color: "blue"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfig("alpha", "beta")

	results := eng.Run(testInput(), cfg)

	if len(results) != 1 {
		t.Fatalf("expected 1 line, got %d", len(results))
	}
	if len(results[0]) != 2 {
		t.Fatalf("expected 2 widgets on line 0, got %d", len(results[0]))
	}

	assertResult(t, results[0][0], "alpha", "A", "green")
	assertResult(t, results[0][1], "beta", "B", "blue")
}

func TestTimedOutWidgetFallsBackToCache(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&slowWidget{name: "slow", delay: 2 * time.Second})

	c := cache.New(t.TempDir())
	cachedOutput := &protocol.WidgetOutput{Text: "cached-slow", Color: "yellow"}
	seedCache(c, "slow", cachedOutput)

	eng := New(reg, c, 100*time.Millisecond)
	cfg := testConfig("slow")

	results := eng.Run(testInput(), cfg)

	if len(results) != 1 || len(results[0]) != 1 {
		t.Fatalf("expected 1 line with 1 widget, got %v", results)
	}

	assertResult(t, results[0][0], "slow", "cached-slow", "yellow")
}

func TestTimedOutWidgetNoCacheIsNil(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&slowWidget{name: "slow", delay: 2 * time.Second})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 100*time.Millisecond)
	cfg := testConfig("slow")

	results := eng.Run(testInput(), cfg)

	if len(results) != 1 || len(results[0]) != 1 {
		t.Fatalf("expected 1 line with 1 widget, got %v", results)
	}

	if results[0][0].Output != nil {
		t.Errorf("expected nil output for timed-out widget with no cache, got %+v", results[0][0].Output)
	}
	if results[0][0].Name != "slow" {
		t.Errorf("expected name 'slow', got %q", results[0][0].Name)
	}
}

func TestErroringWidgetIsNil(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&errorWidget{name: "bad"})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfig("bad")

	results := eng.Run(testInput(), cfg)

	if len(results) != 1 || len(results[0]) != 1 {
		t.Fatalf("expected 1 line with 1 widget, got %v", results)
	}

	if results[0][0].Output != nil {
		t.Errorf("expected nil output for erroring widget, got %+v", results[0][0].Output)
	}
	if results[0][0].Name != "bad" {
		t.Errorf("expected name 'bad', got %q", results[0][0].Name)
	}
}

func TestErroringWidgetFallsBackToCache(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&errorWidget{name: "bad"})

	c := cache.New(t.TempDir())
	cachedOutput := &protocol.WidgetOutput{Text: "cached-bad", Color: "red"}
	seedCache(c, "bad", cachedOutput)

	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfig("bad")

	results := eng.Run(testInput(), cfg)

	if len(results) != 1 || len(results[0]) != 1 {
		t.Fatalf("expected 1 line with 1 widget, got %v", results)
	}

	assertResult(t, results[0][0], "bad", "cached-bad", "red")
}

func TestResultsMaintainLineAndWidgetOrder(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "a", output: &protocol.WidgetOutput{Text: "A", Color: "green"}})
	reg.Register(&fastWidget{name: "b", output: &protocol.WidgetOutput{Text: "B", Color: "blue"}})
	reg.Register(&fastWidget{name: "c", output: &protocol.WidgetOutput{Text: "C", Color: "red"}})
	reg.Register(&fastWidget{name: "d", output: &protocol.WidgetOutput{Text: "D", Color: "yellow"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)

	cfg := multiLineConfig(
		[]string{"b", "a"},
		[]string{"d", "c"},
	)

	results := eng.Run(testInput(), cfg)

	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if len(results[0]) != 2 {
		t.Fatalf("expected 2 widgets on line 0, got %d", len(results[0]))
	}
	if len(results[1]) != 2 {
		t.Fatalf("expected 2 widgets on line 1, got %d", len(results[1]))
	}

	// Line 0: b, a (in config order)
	assertResult(t, results[0][0], "b", "B", "blue")
	assertResult(t, results[0][1], "a", "A", "green")

	// Line 1: d, c (in config order)
	assertResult(t, results[1][0], "d", "D", "yellow")
	assertResult(t, results[1][1], "c", "C", "red")
}

func TestUnregisteredWidgetIsSkipped(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "exists", output: &protocol.WidgetOutput{Text: "E", Color: "green"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfig("exists", "missing", "exists")

	results := eng.Run(testInput(), cfg)

	if len(results) != 1 {
		t.Fatalf("expected 1 line, got %d", len(results))
	}
	// "missing" is not in the registry, so it should be skipped entirely
	// Only the two "exists" entries remain
	if len(results[0]) != 2 {
		t.Fatalf("expected 2 widgets (missing skipped), got %d", len(results[0]))
	}
	assertResult(t, results[0][0], "exists", "E", "green")
	assertResult(t, results[0][1], "exists", "E", "green")
}

func TestSuccessfulWidgetUpdatesCache(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "cached", output: &protocol.WidgetOutput{Text: "fresh", Color: "cyan"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfig("cached")

	eng.Run(testInput(), cfg)

	// Verify the cache was updated
	data, ok := c.Get("cached")
	if !ok {
		t.Fatal("expected cache entry for 'cached' widget")
	}

	var out protocol.WidgetOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to unmarshal cached data: %v", err)
	}
	if out.Text != "fresh" || out.Color != "cyan" {
		t.Errorf("cached output mismatch: got %+v", out)
	}
}

// --- Cache TTL tests ---

func testConfigWithWidgetCfg(widgetName string, widgetCfg map[string]interface{}) *config.Config {
	return &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{widgetName}},
		},
		Widgets: map[string]map[string]interface{}{
			widgetName: widgetCfg,
		},
	}
}

func TestCacheTTLFromWidgetConfig(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "ttl-widget", output: &protocol.WidgetOutput{Text: "ttl", Color: "green"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfigWithWidgetCfg("ttl-widget", map[string]interface{}{
		"cache_ttl": "1s",
	})

	eng.Run(testInput(), cfg)

	// Cache entry should exist immediately after run.
	if _, ok := c.Get("ttl-widget"); !ok {
		t.Fatal("expected cache entry for 'ttl-widget' immediately after run")
	}

	// Wait for the 1s TTL to expire.
	time.Sleep(1100 * time.Millisecond)

	// Cache entry should now be expired.
	if _, ok := c.Get("ttl-widget"); ok {
		t.Fatal("expected cache entry for 'ttl-widget' to be expired after 1s TTL")
	}
}

func TestCacheTTLDefaultWhenNoConfig(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "default-ttl", output: &protocol.WidgetOutput{Text: "ok", Color: "blue"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	// No widget config at all — should use defaultCacheTTL (5 minutes).
	cfg := testConfig("default-ttl")

	eng.Run(testInput(), cfg)

	// Cache entry should exist well within the 5-minute default.
	if _, ok := c.Get("default-ttl"); !ok {
		t.Fatal("expected cache entry for 'default-ttl' with default TTL")
	}
}

func TestCacheTTLInvalidStringUsesDefault(t *testing.T) {
	reg := widget.NewRegistry()
	reg.Register(&fastWidget{name: "bad-ttl", output: &protocol.WidgetOutput{Text: "ok", Color: "red"}})

	c := cache.New(t.TempDir())
	eng := New(reg, c, 500*time.Millisecond)
	cfg := testConfigWithWidgetCfg("bad-ttl", map[string]interface{}{
		"cache_ttl": "not-a-duration",
	})

	eng.Run(testInput(), cfg)

	// Should fall back to default TTL (5 min), so cache should still be valid.
	if _, ok := c.Get("bad-ttl"); !ok {
		t.Fatal("expected cache entry for 'bad-ttl' with default TTL (invalid cache_ttl should be ignored)")
	}
}

func TestGetCacheTTL(t *testing.T) {
	eng := &Engine{}

	tests := []struct {
		name     string
		cfg      map[string]interface{}
		expected time.Duration
	}{
		{"nil config", nil, defaultCacheTTL},
		{"empty config", map[string]interface{}{}, defaultCacheTTL},
		{"valid 10s", map[string]interface{}{"cache_ttl": "10s"}, 10 * time.Second},
		{"valid 2m", map[string]interface{}{"cache_ttl": "2m"}, 2 * time.Minute},
		{"invalid string", map[string]interface{}{"cache_ttl": "bad"}, defaultCacheTTL},
		{"wrong type int", map[string]interface{}{"cache_ttl": 42}, defaultCacheTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eng.getCacheTTL(tt.cfg)
			if got != tt.expected {
				t.Errorf("getCacheTTL(%v) = %v, want %v", tt.cfg, got, tt.expected)
			}
		})
	}
}

// --- assertion helper ---

func assertResult(t *testing.T, wr renderer.WidgetResult, name, text, color string) {
	t.Helper()
	if wr.Name != name {
		t.Errorf("expected name %q, got %q", name, wr.Name)
	}
	if wr.Output == nil {
		t.Fatalf("expected non-nil output for widget %q", name)
	}
	if wr.Output.Text != text {
		t.Errorf("widget %q: expected text %q, got %q", name, text, wr.Output.Text)
	}
	if wr.Output.Color != color {
		t.Errorf("widget %q: expected color %q, got %q", name, color, wr.Output.Color)
	}
}
