package widgets

import (
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestSessionNameWidget_Name(t *testing.T) {
	w := &SessionNameWidget{}
	if w.Name() != "session-name" {
		t.Errorf("expected name %q, got %q", "session-name", w.Name())
	}
}

func TestSessionNameWidget_ReturnsNilWhenEmpty(t *testing.T) {
	w := &SessionNameWidget{}
	input := &protocol.StatusLineInput{}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestSessionNameWidget_ShortNameUnchanged(t *testing.T) {
	assertSessionName(t, "fix-auth-bug", nil, "fix-auth-bug")
}

func TestSessionNameWidget_TruncatesWithDefault(t *testing.T) {
	name := strings.Repeat("a", 40)
	want := strings.Repeat("a", 31) + "…"
	assertSessionName(t, name, nil, want)
}

func TestSessionNameWidget_RespectsMaxLength(t *testing.T) {
	cfg := map[string]interface{}{"max_length": float64(8)}
	assertSessionName(t, "refactor-renderer", cfg, "refacto…")
}

func TestSessionNameWidget_TruncatesOnRunes(t *testing.T) {
	cfg := map[string]interface{}{"max_length": float64(4)}
	assertSessionName(t, "日本語のセッション", cfg, "日本語…")
}

func assertSessionName(t *testing.T, name string, cfg map[string]interface{}, want string) {
	t.Helper()
	w := &SessionNameWidget{}
	input := &protocol.StatusLineInput{SessionName: name}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != want {
		t.Errorf("expected text %q, got %q", want, out.Text)
	}
	if out.Color != "blue" {
		t.Errorf("expected color %q, got %q", "blue", out.Color)
	}
}

func TestSessionNameWidget_StripsControlCharacters(t *testing.T) {
	w := &SessionNameWidget{}
	input := &protocol.StatusLineInput{SessionName: "fix\x1b[31m login\x07 bug\u0085"}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "fix[31m login bug" {
		t.Errorf("expected control characters stripped, got %q", out.Text)
	}
}

func TestSessionNameWidget_ReturnsNilWhenOnlyControlCharacters(t *testing.T) {
	w := &SessionNameWidget{}
	input := &protocol.StatusLineInput{SessionName: "\x1b\x07"}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
