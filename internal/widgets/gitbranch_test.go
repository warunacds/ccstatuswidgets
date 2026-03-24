package widgets

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestGitBranchWidget_Name(t *testing.T) {
	w := &GitBranchWidget{}
	if w.Name() != "git-branch" {
		t.Errorf("expected name %q, got %q", "git-branch", w.Name())
	}
}

func TestGitBranchWidget_ReturnsBranchInYellow(t *testing.T) {
	// Create a temporary git repo.
	tmpDir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "checkout", "-b", "feature-xyz"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}
	// Need at least one commit for symbolic-ref to work.
	dummyFile := filepath.Join(tmpDir, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	w := &GitBranchWidget{}
	input := &protocol.StatusLineInput{
		Workspace: protocol.WorkspaceInfo{
			CurrentDir: tmpDir,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "(feature-xyz)" {
		t.Errorf("expected text %q, got %q", "(feature-xyz)", out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}

func TestGitBranchWidget_ReturnsNilWhenNotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	w := &GitBranchWidget{}
	input := &protocol.StatusLineInput{
		Workspace: protocol.WorkspaceInfo{
			CurrentDir: tmpDir,
		},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
