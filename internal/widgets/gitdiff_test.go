package widgets

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestGitDiffWidget_Name(t *testing.T) {
	w := &GitDiffWidget{}
	if w.Name() != "git-diff" {
		t.Errorf("expected name %q, got %q", "git-diff", w.Name())
	}
}

func TestGitDiffWidget_ShowsStagedFileCount(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a git repo with an initial commit.
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

	// Create and commit a dummy file so we have a HEAD.
	dummyFile := filepath.Join(tmpDir, "dummy.txt")
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

	// Stage some new files.
	for _, name := range []string{"a.go", "b.go", "c.go", "d.go"} {
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	w := &GitDiffWidget{}
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
	if out.Text != "staged: 4 files" {
		t.Errorf("expected text %q, got %q", "staged: 4 files", out.Text)
	}
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}

func TestGitDiffWidget_ReturnsNilWhenNoStagedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a git repo with an initial commit but no staged changes.
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

	dummyFile := filepath.Join(tmpDir, "dummy.txt")
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

	w := &GitDiffWidget{}
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
		t.Errorf("expected nil output when no staged files, got %+v", out)
	}
}

func TestGitDiffWidget_ReturnsNilWhenNotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	w := &GitDiffWidget{}
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

func TestGitDiffWidget_ColorIsGreen(t *testing.T) {
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

	dummyFile := filepath.Join(tmpDir, "dummy.txt")
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

	// Stage one file.
	f := filepath.Join(tmpDir, "new.go")
	if err := os.WriteFile(f, []byte("package x"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "new.go")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	w := &GitDiffWidget{}
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
	if out.Color != "green" {
		t.Errorf("expected color %q, got %q", "green", out.Color)
	}
}
