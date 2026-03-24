package cli

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

// RunAdd adds a widget to the config layout.
//
//	ccw add <widget_name>            — appends to last row
//	ccw add <widget_name> -r <row>   — appends to specified row (1-indexed)
//	ccw add <widget_name> -r <row> -c <col> — inserts at specific position
func RunAdd(args []string, configDir string, registry *widget.Registry) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ccw add <widget_name> [-r row] [-c col] [--rc row:col]")
	}

	// Parse widget name and flags manually.
	widgetName := args[0]
	row := 0
	col := 0

	for i := 1; i < len(args); i++ {
		switch {
		case args[i] == "-r":
			if i+1 >= len(args) {
				return fmt.Errorf("missing value for -r flag")
			}
			i++
			v, err := strconv.Atoi(args[i])
			if err != nil || v < 1 {
				return fmt.Errorf("invalid row number: %s", args[i])
			}
			row = v
		case args[i] == "-c":
			if i+1 >= len(args) {
				return fmt.Errorf("missing value for -c flag")
			}
			i++
			v, err := strconv.Atoi(args[i])
			if err != nil || v < 1 {
				return fmt.Errorf("invalid column number: %s", args[i])
			}
			col = v
		case strings.HasPrefix(args[i], "--rc"):
			val := ""
			if args[i] == "--rc" {
				if i+1 >= len(args) {
					return fmt.Errorf("missing value for --rc flag")
				}
				i++
				val = args[i]
			} else if strings.HasPrefix(args[i], "--rc=") {
				val = args[i][5:]
			} else {
				return fmt.Errorf("unknown flag: %s", args[i])
			}
			parts := strings.SplitN(val, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --rc format, expected row:col (e.g., --rc 2:4)")
			}
			rv, err := strconv.Atoi(parts[0])
			if err != nil || rv < 1 {
				return fmt.Errorf("invalid row in --rc: %s", parts[0])
			}
			cv, err := strconv.Atoi(parts[1])
			if err != nil || cv < 1 {
				return fmt.Errorf("invalid col in --rc: %s", parts[1])
			}
			row = rv
			col = cv
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	// Validate widget name exists in registry.
	if _, ok := registry.Get(widgetName); !ok {
		return fmt.Errorf("unknown widget: %s", widgetName)
	}

	// Load config.
	configPath := filepath.Join(configDir, "config.json")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if widget is already in the config.
	for i, line := range cfg.Lines {
		for j, w := range line.Widgets {
			if w == widgetName {
				fmt.Printf("%s is already enabled (row %d, col %d)\n", widgetName, i+1, j+1)
				return nil
			}
		}
	}

	// Determine target row.
	targetRow := row
	if targetRow == 0 {
		// Default: last row.
		targetRow = len(cfg.Lines)
		if targetRow == 0 {
			targetRow = 1
		}
	}

	// Create rows up to targetRow if needed.
	for len(cfg.Lines) < targetRow {
		cfg.Lines = append(cfg.Lines, config.LineConfig{Widgets: []string{}})
	}

	// Insert widget.
	lineIdx := targetRow - 1
	if col == 0 {
		// Append to end of row.
		cfg.Lines[lineIdx].Widgets = append(cfg.Lines[lineIdx].Widgets, widgetName)
		col = len(cfg.Lines[lineIdx].Widgets)
	} else {
		// Insert at specific position (1-indexed).
		widgets := cfg.Lines[lineIdx].Widgets
		if col > len(widgets) {
			// Append to end.
			cfg.Lines[lineIdx].Widgets = append(widgets, widgetName)
			col = len(cfg.Lines[lineIdx].Widgets)
		} else {
			// Insert at position.
			newWidgets := make([]string, 0, len(widgets)+1)
			newWidgets = append(newWidgets, widgets[:col-1]...)
			newWidgets = append(newWidgets, widgetName)
			newWidgets = append(newWidgets, widgets[col-1:]...)
			cfg.Lines[lineIdx].Widgets = newWidgets
		}
	}

	// Save config.
	if err := config.Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added %s to row %d, col %d\n", widgetName, targetRow, col)
	return nil
}
