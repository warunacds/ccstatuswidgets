package widgets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestStocksWidget_Name(t *testing.T) {
	w := &StocksWidget{}
	if w.Name() != "stocks" {
		t.Errorf("expected name %q, got %q", "stocks", w.Name())
	}
}

func TestStocksWidget_FormatsStockChanges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":150.0,"previousClose":148.5}}]}}`))
	}))
	defer srv.Close()

	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"symbols":  []interface{}{"AAPL"},
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.Contains(out.Text, "AAPL") {
		t.Errorf("expected output to contain symbol AAPL, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "+1.0%") {
		t.Errorf("expected output to contain +1.0%%, got %q", out.Text)
	}
}

func TestStocksWidget_GreenANSIForPositiveChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// price up: 150 from 148.5 = +1.01%
		w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":150.0,"previousClose":148.5}}]}}`))
	}))
	defer srv.Close()

	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"symbols":  []interface{}{"AAPL"},
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	green := "\033[0;32m"
	if !strings.Contains(out.Text, green) {
		t.Errorf("expected green ANSI code in output, got %q", out.Text)
	}
	if out.Color != "" {
		t.Errorf("expected empty color, got %q", out.Color)
	}
}

func TestStocksWidget_RedANSIForNegativeChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// price down: 140 from 148.5 = -5.72%
		w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":140.0,"previousClose":148.5}}]}}`))
	}))
	defer srv.Close()

	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"symbols":  []interface{}{"TSLA"},
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	red := "\033[0;31m"
	if !strings.Contains(out.Text, red) {
		t.Errorf("expected red ANSI code in output, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "TSLA") {
		t.Errorf("expected output to contain symbol TSLA, got %q", out.Text)
	}
}

func TestStocksWidget_MultipleSymbolsSeparatedBySpaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Serve different responses based on the requested symbol path
		if strings.Contains(r.URL.Path, "AAPL") {
			w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":150.0,"previousClose":148.5}}]}}`))
		} else if strings.Contains(r.URL.Path, "TSLA") {
			w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":140.0,"previousClose":148.5}}]}}`))
		}
	}))
	defer srv.Close()

	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"symbols":  []interface{}{"AAPL", "TSLA"},
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.Contains(out.Text, "AAPL") {
		t.Errorf("expected output to contain AAPL, got %q", out.Text)
	}
	if !strings.Contains(out.Text, "TSLA") {
		t.Errorf("expected output to contain TSLA, got %q", out.Text)
	}
	// Verify they're separated by space - find AAPL section end and TSLA section start
	aaplIdx := strings.Index(out.Text, "AAPL")
	tslaIdx := strings.Index(out.Text, "TSLA")
	if aaplIdx >= tslaIdx {
		t.Errorf("expected AAPL before TSLA in output, got %q", out.Text)
	}
}

func TestStocksWidget_ReturnsNilWhenNoSymbols(t *testing.T) {
	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}

	// No symbols in cfg
	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}

	// Empty symbols list
	cfg := map[string]interface{}{
		"symbols": []interface{}{},
	}
	out, err = w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output for empty symbols, got %+v", out)
	}
}

func TestStocksWidget_ReturnsNilWhenHTTPFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"symbols":  []interface{}{"AAPL"},
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output when HTTP fails, got %+v", out)
	}
}

func TestStocksWidget_PercentFormatting(t *testing.T) {
	// Verify the exact percent format: 1 decimal place
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// (153.123 - 150.0) / 150.0 * 100 = 2.082%
		w.Write([]byte(`{"chart":{"result":[{"meta":{"regularMarketPrice":153.123,"previousClose":150.0}}]}}`))
	}))
	defer srv.Close()

	w := &StocksWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"symbols":  []interface{}{"GOOG"},
		"base_url": srv.URL,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	expected := fmt.Sprintf("\033[0;32mGOOG +2.1%%\033[0m")
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
}
