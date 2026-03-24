package cli

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

// RunConfigEdit opens the ccstatuswidgets config file in the user's editor.
func RunConfigEdit() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	configPath := filepath.Join(config.ConfigDir(), "config.json")
	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
