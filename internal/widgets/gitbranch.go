package widgets

import (
	"os"
	"os/exec"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// GitBranchWidget displays the current git branch name.
type GitBranchWidget struct{}

func (w *GitBranchWidget) Name() string {
	return "git-branch"
}

func (w *GitBranchWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	dir := input.Workspace.CurrentDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil
		}
		dir = cwd
	}

	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.Output()
	if err != nil {
		// Not in a git repo or detached HEAD — return nil.
		return nil, nil
	}

	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return nil, nil
	}

	return &protocol.WidgetOutput{
		Text:  "(" + branch + ")",
		Color: "yellow",
	}, nil
}
