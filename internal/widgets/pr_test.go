package widgets

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestPRWidget_Name(t *testing.T) {
	w := &PRWidget{}
	if w.Name() != "pr" {
		t.Errorf("expected name %q, got %q", "pr", w.Name())
	}
}

func TestPRWidget_ReturnsNilWhenNoPR(t *testing.T) {
	w := &PRWidget{}
	input := &protocol.StatusLineInput{}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPRWidget_Approved(t *testing.T) {
	assertPR(t, &protocol.PRInfo{Number: 1234, ReviewState: "approved"}, "PR #1234 ✓", "green")
}

func TestPRWidget_ChangesRequested(t *testing.T) {
	assertPR(t, &protocol.PRInfo{Number: 1234, ReviewState: "changes_requested"}, "PR #1234 ✗", "red")
}

func TestPRWidget_Pending(t *testing.T) {
	assertPR(t, &protocol.PRInfo{Number: 1234, ReviewState: "pending"}, "PR #1234 ○", "yellow")
}

func TestPRWidget_Draft(t *testing.T) {
	assertPR(t, &protocol.PRInfo{Number: 1234, ReviewState: "draft"}, "PR #1234 draft", "dim")
}

func TestPRWidget_NoState(t *testing.T) {
	assertPR(t, &protocol.PRInfo{Number: 1234}, "PR #1234", "cyan")
}

func TestPRWidget_GitLabMR(t *testing.T) {
	assertPR(t, &protocol.PRInfo{Number: 42, ReviewState: "approved", Kind: "mr"}, "MR #42 ✓", "green")
}

func assertPR(t *testing.T, pr *protocol.PRInfo, wantText, wantColor string) {
	t.Helper()
	w := &PRWidget{}
	input := &protocol.StatusLineInput{PR: pr}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != wantText {
		t.Errorf("expected text %q, got %q", wantText, out.Text)
	}
	if out.Color != wantColor {
		t.Errorf("expected color %q, got %q", wantColor, out.Color)
	}
}
