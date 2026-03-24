package cli

import (
	"fmt"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

// RunRemove removes a widget from the config layout.
//
//	ccw remove <widget_name>
func RunRemove(args []string, configDir string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ccw remove <widget_name>")
	}

	widgetName := args[0]

	// Load config.
	configPath := filepath.Join(configDir, "config.json")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find and remove the widget.
	found := false
	foundRow := 0
	for i, line := range cfg.Lines {
		for j, w := range line.Widgets {
			if w == widgetName {
				cfg.Lines[i].Widgets = append(line.Widgets[:j], line.Widgets[j+1:]...)
				found = true
				foundRow = i + 1
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		fmt.Printf("%s is not enabled\n", widgetName)
		return nil
	}

	// Remove empty lines.
	var nonEmpty []config.LineConfig
	for _, line := range cfg.Lines {
		if len(line.Widgets) > 0 {
			nonEmpty = append(nonEmpty, line)
		}
	}
	cfg.Lines = nonEmpty

	// Save config.
	if err := config.Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Removed %s from row %d\n", widgetName, foundRow)
	return nil
}
