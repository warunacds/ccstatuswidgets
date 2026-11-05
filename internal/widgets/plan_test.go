package widgets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func writePlanState(t *testing.T, content string) map[string]interface{} {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".claude.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}
	return map[string]interface{}{"state_file": path}
}

func TestPlanWidget_Name(t *testing.T) {
	w := &PlanWidget{}
	if w.Name() != "plan" {
		t.Errorf("expected name %q, got %q", "plan", w.Name())
	}
}

func TestPlanWidget_ShowsPlanDisplayName(t *testing.T) {
	cfg := writePlanState(t, `{"oauthAccount":{"emailAddress":"x@y.z","planDisplayName":"Ultra","organizationType":"claude_max"}}`)

	out, err := (&PlanWidget{}).Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "Ultra" {
		t.Errorf("expected text %q, got %q", "Ultra", out.Text)
	}
	if out.Color != "magenta" {
		t.Errorf("expected color %q, got %q", "magenta", out.Color)
	}
}

func TestPlanWidget_ReturnsNilWhenNoOAuthAccount(t *testing.T) {
	cfg := writePlanState(t, `{"numStartups":3}`)

	out, err := (&PlanWidget{}).Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPlanWidget_ReturnsNilWhenFieldMissing(t *testing.T) {
	cfg := writePlanState(t, `{"oauthAccount":{"emailAddress":"x@y.z"}}`)

	out, err := (&PlanWidget{}).Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPlanWidget_ReturnsNilWhenFileMissing(t *testing.T) {
	cfg := map[string]interface{}{"state_file": filepath.Join(t.TempDir(), "missing.json")}

	out, err := (&PlanWidget{}).Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPlanWidget_ReturnsNilWhenInvalidJSON(t *testing.T) {
	cfg := writePlanState(t, `{not json`)

	out, err := (&PlanWidget{}).Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
