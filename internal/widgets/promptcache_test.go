package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestPromptCacheWidget_Name(t *testing.T) {
	w := &PromptCacheWidget{}
	if w.Name() != "prompt-cache" {
		t.Errorf("expected name %q, got %q", "prompt-cache", w.Name())
	}
}

func TestPromptCacheWidget_ReturnsNilWhenAbsent(t *testing.T) {
	w := &PromptCacheWidget{}
	input := &protocol.StatusLineInput{}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPromptCacheWidget_ReturnsNilWhenCachingNotObserved(t *testing.T) {
	w := &PromptCacheWidget{}
	input := &protocol.StatusLineInput{
		PromptCache: &protocol.PromptCache{Warm: true, CachingObserved: false, HitRatio: 0.9},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPromptCacheWidget_WarmInGreen(t *testing.T) {
	w := &PromptCacheWidget{}
	input := &protocol.StatusLineInput{
		PromptCache: &protocol.PromptCache{Warm: true, CachingObserved: true, HitRatio: 0.91},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "cache 91%" {
		t.Errorf("expected text %q, got %q", "cache 91%", out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestPromptCacheWidget_ColdInDim(t *testing.T) {
	w := &PromptCacheWidget{}
	input := &protocol.StatusLineInput{
		PromptCache: &protocol.PromptCache{Warm: false, CachingObserved: true, HitRatio: 0.91},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "cache 91% cold" {
		t.Errorf("expected text %q, got %q", "cache 91% cold", out.Text)
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}

func TestPromptCacheWidget_RoundsPercentage(t *testing.T) {
	tests := []struct {
		ratio float64
		want  string
	}{
		{0.915, "cache 92%"},
		{0.004, "cache 0%"},
		{1.0, "cache 100%"},
	}
	w := &PromptCacheWidget{}
	for _, tt := range tests {
		input := &protocol.StatusLineInput{
			PromptCache: &protocol.PromptCache{Warm: true, CachingObserved: true, HitRatio: tt.ratio},
		}
		out, err := w.Render(input, nil)
		if err != nil {
			t.Fatalf("ratio %v: unexpected error: %v", tt.ratio, err)
		}
		if out == nil {
			t.Fatalf("ratio %v: expected non-nil output", tt.ratio)
		}
		if out.Text != tt.want {
			t.Errorf("ratio %v: expected text %q, got %q", tt.ratio, tt.want, out.Text)
		}
	}
}
