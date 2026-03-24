package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

// pluginManifest represents the plugin.json file in a plugin directory.
type pluginManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Entry       string `json:"entry"`
}

// RunPlugin handles the "plugin" subcommand.
//
//	ccw plugin add <github-url>
//	ccw plugin list
//	ccw plugin remove <name>
//	ccw plugin update <name>
//	ccw plugin update --all
func RunPlugin(args []string, pluginsDir string, builtinNames []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ccw plugin <add|list|remove|update>")
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: ccw plugin add <github-url>")
		}
		return pluginAdd(args[1], pluginsDir, builtinNames)
	case "list":
		return pluginList(pluginsDir)
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: ccw plugin remove <name>")
		}
		return pluginRemove(args[1], pluginsDir)
	case "update":
		if len(args) < 2 {
			return fmt.Errorf("usage: ccw plugin update <name|--all>")
		}
		if args[1] == "--all" {
			return pluginUpdateAll(pluginsDir)
		}
		return pluginUpdate(args[1], pluginsDir)
	default:
		return fmt.Errorf("unknown plugin command: %s", args[0])
	}
}

func pluginAdd(url string, pluginsDir string, builtinNames []string) error {
	// Derive repo name from URL.
	repoName := filepath.Base(url)
	repoName = strings.TrimSuffix(repoName, ".git")

	cloneDir := filepath.Join(pluginsDir, repoName)

	// Clone the repository.
	cmd := exec.Command("git", "clone", url, cloneDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s\n%s", err, output)
	}

	// Validate plugin.json exists.
	manifestPath := filepath.Join(cloneDir, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		// Clean up the cloned directory.
		os.RemoveAll(cloneDir)
		return fmt.Errorf("plugin.json not found in repository — not a valid plugin")
	}

	var manifest pluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		os.RemoveAll(cloneDir)
		return fmt.Errorf("plugin.json is not valid JSON: %w", err)
	}

	if manifest.Name == "" {
		os.RemoveAll(cloneDir)
		return fmt.Errorf("plugin.json is missing required 'name' field")
	}

	// Check for name conflict with built-in widgets.
	for _, builtin := range builtinNames {
		if manifest.Name == builtin {
			os.RemoveAll(cloneDir)
			return fmt.Errorf("plugin name %q conflicts with built-in widget", manifest.Name)
		}
	}

	fmt.Printf("Plugin %q (v%s) installed successfully.\n", manifest.Name, manifest.Version)
	return nil
}

func pluginList(pluginsDir string) error {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plugins installed.")
			return nil
		}
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	var plugins []pluginManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(pluginsDir, entry.Name(), "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var manifest pluginManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		plugins = append(plugins, manifest)
	}

	if len(plugins) == 0 {
		fmt.Println("No plugins installed.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION")
	for _, p := range plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Version, p.Description)
	}
	w.Flush()

	return nil
}

func pluginRemove(name string, pluginsDir string) error {
	dir, err := findPluginDir(name, pluginsDir)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove plugin directory: %w", err)
	}

	fmt.Printf("Plugin %q removed.\n", name)
	return nil
}

func pluginUpdate(name string, pluginsDir string) error {
	dir, err := findPluginDir(name, pluginsDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "-C", dir, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed for %q: %s\n%s", name, err, output)
	}

	fmt.Printf("Plugin %q updated.\n", name)
	return nil
}

func pluginUpdateAll(pluginsDir string) error {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plugins installed.")
			return nil
		}
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	updated := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(pluginsDir, entry.Name())
		manifestPath := filepath.Join(dir, "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var manifest pluginManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		cmd := exec.Command("git", "-C", dir, "pull")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: git pull failed for %q: %s\n%s", manifest.Name, err, output)
			continue
		}
		fmt.Printf("Plugin %q updated.\n", manifest.Name)
		updated++
	}

	if updated == 0 {
		fmt.Println("No plugins to update.")
	}

	return nil
}

// findPluginDir searches pluginsDir for a plugin whose plugin.json has the given name.
func findPluginDir(name string, pluginsDir string) (string, error) {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(pluginsDir, entry.Name())
		manifestPath := filepath.Join(dir, "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var manifest pluginManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		if manifest.Name == name {
			return dir, nil
		}
	}

	return "", fmt.Errorf("plugin %q not found", name)
}
