package widgets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestDirectoryWidget_Name(t *testing.T) {
	w := &DirectoryWidget{}
	if w.Name() != "directory" {
		t.Errorf("expected name %q, got %q", "directory", w.Name())
	}
}

func TestDirectoryWidget_ReturnsBasenameInCyan(t *testing.T) {
	w := &DirectoryWidget{}
	input := &protocol.StatusLineInput{
		Workspace: protocol.WorkspaceInfo{
			CurrentDir: "/home/user/projects/myapp",
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "myapp" {
		t.Errorf("expected text %q, got %q", "myapp", out.Text)
	}
	if out.Color != "cyan" {
		t.Errorf("expected color %q, got %q", "cyan", out.Color)
	}
}

func TestDirectoryWidget_FallsBackToPWD(t *testing.T) {
	w := &DirectoryWidget{}
	input := &protocol.StatusLineInput{
		Workspace: protocol.WorkspaceInfo{
			CurrentDir: "",
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}

	cwd, _ := os.Getwd()
	expected := filepath.Base(cwd)
	if out.Text != expected {
		t.Errorf("expected text %q (from PWD), got %q", expected, out.Text)
	}
	if out.Color != "cyan" {
		t.Errorf("expected color %q, got %q", "cyan", out.Color)
	}
}
