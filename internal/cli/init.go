package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

// RunInit creates the ccstatuswidgets configuration directory structure,
// writes a default config, and patches Claude Code's settings.json.
func RunInit() error {
	configDir := config.ConfigDir()

	// Create directory structure.
	dirs := []string{
		configDir,
		filepath.Join(configDir, "cache"),
		filepath.Join(configDir, "plugins"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Write default config.
	configPath := filepath.Join(configDir, "config.json")
	cfg := config.Default()
	if err := config.Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Patch Claude Code settings.json if it exists.
	if err := patchClaudeSettings(); err != nil {
		// Non-fatal — print warning and continue.
		fmt.Fprintf(os.Stderr, "warning: could not patch Claude Code settings: %v\n", err)
	}

	fmt.Println("ccstatuswidgets initialized successfully!")
	fmt.Println()
	fmt.Println("Created:")
	fmt.Printf("  %s\n", configPath)
	fmt.Printf("  %s\n", filepath.Join(configDir, "cache"))
	fmt.Printf("  %s\n", filepath.Join(configDir, "plugins"))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'ccw doctor' to verify your setup")
	fmt.Println("  2. Run 'ccw preview' to see a sample status line")
	fmt.Println("  3. Start Claude Code — the status line should appear automatically")

	return nil
}

// patchClaudeSettings reads ~/.claude/settings.json and adds or updates
// the statusLine key to point to the ccw binary. If the file does not
// exist, it is skipped silently.
func patchClaudeSettings() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil // skip silently
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // skip silently
		}
		return err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("invalid JSON in %s: %w", settingsPath, err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine binary path: %w", err)
	}

	settings["statusLine"] = map[string]interface{}{
		"type":    "command",
		"command": exePath,
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}
