package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// PluginManifest describes a plugin's metadata, loaded from plugin.json.
type PluginManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Entry       string `json:"entry"`
	CacheTTL    string `json:"cache_ttl"`
	Interpreter string `json:"interpreter"`
}

// ExternalWidget wraps an external plugin as a Widget.
type ExternalWidget struct {
	manifest PluginManifest
	dir      string
}

// Name returns the plugin's name from its manifest.
func (w *ExternalWidget) Name() string { return w.manifest.Name }

// Render executes the plugin's entry script, piping StatusLineInput as JSON to
// stdin and reading WidgetOutput JSON from stdout.
func (w *ExternalWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	var cmd *exec.Cmd
	if w.manifest.Interpreter != "" {
		cmd = exec.Command(w.manifest.Interpreter, w.manifest.Entry)
	} else {
		cmd = exec.Command(filepath.Join(w.dir, w.manifest.Entry))
	}
	cmd.Dir = w.dir

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("plugin %s: failed to marshal input: %w", w.manifest.Name, err)
	}

	cmd.Stdin = bytes.NewReader(inputJSON)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("plugin %s: execution failed: %w", w.manifest.Name, err)
	}

	var result protocol.WidgetOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid JSON output: %w", w.manifest.Name, err)
	}

	return &result, nil
}

// DiscoverPlugins scans pluginsDir for subdirectories containing a valid
// plugin.json and returns an ExternalWidget for each.
func DiscoverPlugins(pluginsDir string) ([]*ExternalWidget, error) {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("reading plugins directory: %w", err)
	}

	var plugins []*ExternalWidget
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(pluginsDir, entry.Name())
		manifestPath := filepath.Join(dir, "plugin.json")

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			// No plugin.json — skip
			continue
		}

		var manifest PluginManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			// Invalid JSON — skip
			continue
		}

		plugins = append(plugins, &ExternalWidget{
			manifest: manifest,
			dir:      dir,
		})
	}

	return plugins, nil
}
