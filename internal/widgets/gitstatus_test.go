package widgets

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestGitStatusWidget_Name(t *testing.T) {
	w := &GitStatusWidget{}
	if w.Name() != "git-status" {
		t.Errorf("expected name %q, got %q", "git-status", w.Name())
	}
}

// initTestRepo creates a temp git repo with one committed file and returns the dir path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	// Create and commit a file so the repo has at least one commit.
	dummyFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(dummyFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}
	return tmpDir
}

func TestGitStatusWidget_CleanRepo(t *testing.T) {
	tmpDir := initTestRepo(t)

	w := &GitStatusWidget{}
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
	if out.Text != "\u2713" {
		t.Errorf("expected text %q, got %q", "\u2713", out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestGitStatusWidget_DirtyRepoModified(t *testing.T) {
	tmpDir := initTestRepo(t)

	// Modify the tracked file.
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}

	w := &GitStatusWidget{}
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
	if out.Text != "\u270e 1M" {
		t.Errorf("expected text %q, got %q", "\u270e 1M", out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}

func TestGitStatusWidget_UntrackedFiles(t *testing.T) {
	tmpDir := initTestRepo(t)

	// Add an untracked file.
	if err := os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	w := &GitStatusWidget{}
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
	if out.Text != "\u270e 1U" {
		t.Errorf("expected text %q, got %q", "\u270e 1U", out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}

func TestGitStatusWidget_ReturnsNilWhenNotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	w := &GitStatusWidget{}
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

func TestGitStatusWidget_CombinedCounts(t *testing.T) {
	tmpDir := initTestRepo(t)

	// Create and commit a file we will later delete (D).
	delFile := filepath.Join(tmpDir, "todelete.txt")
	if err := os.WriteFile(delFile, []byte("delete me"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "todelete.txt"},
		{"git", "commit", "-m", "add todelete"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	// Now stage the deletion.
	cmd := exec.Command("git", "rm", "todelete.txt")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git rm failed: %v\n%s", err, out)
	}

	// Modify tracked file (M) — unstaged working tree change.
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a new file to the index (A) — staged addition.
	addedFile := filepath.Join(tmpDir, "added.txt")
	if err := os.WriteFile(addedFile, []byte("added"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "add", "added.txt")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	// Add an untracked file (U).
	if err := os.WriteFile(filepath.Join(tmpDir, "untracked.txt"), []byte("u"), 0644); err != nil {
		t.Fatal(err)
	}

	w := &GitStatusWidget{}
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
	// Expect: modified=1, added=1, deleted=1, untracked=1
	expected := "\u270e 1M 1A 1D 1U"
	if out.Text != expected {
		t.Errorf("expected text %q, got %q", expected, out.Text)
	}
	if out.Color != "yellow" {
		t.Errorf("expected color %q, got %q", "yellow", out.Color)
	}
}
