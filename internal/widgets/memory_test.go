package widgets

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestMemoryWidget_Name(t *testing.T) {
	w := &MemoryWidget{}
	if w.Name() != "memory" {
		t.Errorf("expected name %q, got %q", "memory", w.Name())
	}
}

func TestMemoryWidget_ReturnsMemoryInMBFormat(t *testing.T) {
	w := &MemoryWidget{}
	input := &protocol.StatusLineInput{}
	// Use current process pid since getppid in test returns the test runner.
	cfg := map[string]interface{}{
		"pid": os.Getpid(),
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if !strings.HasSuffix(out.Text, "MB") {
		t.Errorf("expected text to end with MB, got %q", out.Text)
	}
	// Verify the number part is a valid integer.
	numStr := strings.TrimSuffix(out.Text, "MB")
	n, err := strconv.Atoi(numStr)
	if err != nil {
		t.Errorf("expected numeric prefix, got %q: %v", numStr, err)
	}
	if n <= 0 {
		t.Errorf("expected positive memory value, got %d", n)
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}

func TestMemoryWidget_ReturnsNilWhenProcessNotFound(t *testing.T) {
	w := &MemoryWidget{}
	input := &protocol.StatusLineInput{}
	// Use an invalid pid that won't exist.
	cfg := map[string]interface{}{
		"pid": 9999999,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
