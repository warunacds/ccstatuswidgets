package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

// RunList shows all available widgets (built-in + installed plugins) with their
// enabled status and position in the config layout.
func RunList(configDir string, registry *widget.Registry, pluginsDir string) error {
	// Load config.
	cfg, err := config.Load(filepath.Join(configDir, "config.json"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build a map of widget name -> (row, col) from config lines.
	type position struct {
		row int
		col int
	}
	positions := make(map[string]position)
	for i, line := range cfg.Lines {
		for j, w := range line.Widgets {
			positions[w] = position{row: i + 1, col: j + 1}
		}
	}

	// Get built-in widget names and sort them.
	builtinNames := registry.Names()
	sort.Strings(builtinNames)

	fmt.Println("Built-in Widgets:")
	for _, name := range builtinNames {
		if pos, ok := positions[name]; ok {
			fmt.Printf("  ✓ %-18s row %d, col %d\n", name, pos.row, pos.col)
		} else {
			fmt.Printf("  ✗ %-18s disabled\n", name)
		}
	}

	// Scan plugins directory.
	fmt.Println()
	fmt.Println("Plugins:")

	pluginNames, err := discoverPluginNames(pluginsDir)
	if err != nil || len(pluginNames) == 0 {
		fmt.Println("  (none installed)")
		return nil
	}

	sort.Strings(pluginNames)
	for _, name := range pluginNames {
		if pos, ok := positions[name]; ok {
			fmt.Printf("  ✓ %-18s row %d, col %d\n", name, pos.row, pos.col)
		} else {
			fmt.Printf("  ✗ %-18s disabled\n", name)
		}
	}

	return nil
}

// discoverPluginNames reads plugin.json from each subdirectory in pluginsDir
// and returns the plugin names.
func discoverPluginNames(pluginsDir string) ([]string, error) {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(pluginsDir, entry.Name(), "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var manifest struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		if manifest.Name != "" {
			names = append(names, manifest.Name)
		}
	}
	return names, nil
}
