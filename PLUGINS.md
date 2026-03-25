# Building Plugins for ccstatuswidgets

A complete guide to creating, testing, and publishing custom widgets for ccstatuswidgets.

---

## How Plugins Work

A plugin is a standalone executable that ccstatuswidgets runs during each status line render. The engine pipes JSON data to your plugin's stdin, and your plugin writes a JSON response to stdout. That's the entire contract.

```
┌──────────────┐     stdin (JSON)      ┌──────────────┐
│  ccw engine  │ ───────────────────▸  │  your plugin │
│              │                       │              │
│              │  ◂───────────────────  │              │
└──────────────┘     stdout (JSON)     └──────────────┘
```

Plugins can be written in **any language** — shell scripts, Python, Go, Ruby, Node.js, Rust — as long as they can read stdin and write to stdout.

---

## Quick Start

### 1. Create a directory

```bash
mkdir my-ccw-plugin
cd my-ccw-plugin
```

### 2. Create `plugin.json`

```json
{
  "name": "my-widget",
  "version": "1.0.0",
  "description": "Shows something useful",
  "entry": "widget.sh",
  "cache_ttl": "5m",
  "interpreter": "sh"
}
```

### 3. Create your widget script

```sh
#!/bin/sh
# Read and discard stdin (required — the engine pipes data to us)
cat > /dev/null

# Your logic here
printf '{"text": "hello world", "color": "cyan"}'
```

### 4. Test it

```bash
echo '{}' | sh widget.sh
# Output: {"text": "hello world", "color": "cyan"}
```

### 5. Make it a git repo

```bash
git init
git add .
git commit -m "Initial commit"
```

### 6. Install and enable

```bash
# Install from local path (for development)
cp -r . ~/.ccstatuswidgets/plugins/my-ccw-plugin/

# Or publish to GitHub and install remotely
ccw plugin add github.com/yourname/my-ccw-plugin

# Enable on the status line
ccw add my-widget
```

---

## Plugin Manifest (`plugin.json`)

Every plugin must have a `plugin.json` at the repository root.

```json
{
  "name": "my-widget",
  "version": "1.0.0",
  "description": "One-line description of what it shows",
  "entry": "widget.sh",
  "cache_ttl": "5m",
  "interpreter": "sh"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Widget identifier. Used in `config.json` lines and `ccw add`. Must be unique — cannot conflict with built-in widget names. |
| `version` | Yes | Semver version string. |
| `description` | Yes | One-line description shown in `ccw plugin list`. |
| `entry` | Yes | Relative path to the entry point file. |
| `cache_ttl` | No | How long to cache this plugin's output. Default: `5m`. Accepts Go duration strings: `30s`, `5m`, `2h`, `24h`. |
| `interpreter` | No | Runtime to execute the entry file. If omitted, the entry file must be directly executable (`chmod +x`). Common values: `sh`, `python3`, `node`, `ruby`. |

### Name Rules

- Use lowercase with hyphens: `my-widget`, `weather-extended`, `cpu-temp`
- Cannot conflict with built-in names: `model`, `effort`, `directory`, `git-branch`, `context-bar`, `tokens`, `session-time`, `usage-5h`, `usage-7d`, `lines-changed`, `git-status`, `cost`, `memory`, `weather`, `now-playing`, `flight`, `cricket`, `stocks`, `hackernews`, `moon`, `pomodoro`, `total-tokens`, `git-diff`
- `ccw plugin add` will reject plugins with conflicting names

---

## Input / Output Protocol

### Input (stdin)

Your plugin receives a `StatusLineInput` JSON object on stdin. This is the same data Claude Code provides to ccstatuswidgets:

```json
{
  "model": {
    "id": "claude-opus-4-6",
    "display_name": "Opus 4.6 (1M context)"
  },
  "workspace": {
    "current_dir": "/Users/you/project",
    "project_dir": "/Users/you/project"
  },
  "context_window": {
    "used_percentage": 35.2,
    "remaining_percentage": 64.8,
    "total_input_tokens": 334100,
    "total_output_tokens": 596400,
    "context_window_size": 1000000
  },
  "rate_limits": {
    "five_hour": {
      "used_percentage": 19.5,
      "resets_at": 1711300000
    },
    "seven_day": {
      "used_percentage": 31.2,
      "resets_at": 1711900000
    }
  },
  "cost": {
    "total_cost_usd": 105.81,
    "total_lines_added": 19361,
    "total_lines_removed": 761
  },
  "session_id": "abc123",
  "version": "1.0.85"
}
```

**Important:** You must read stdin even if you don't use it. If you don't consume stdin, the engine's pipe may block. The simplest approach:

```sh
cat > /dev/null    # shell
```
```python
import sys
sys.stdin.read()   # Python
```

### Output (stdout)

Your plugin must write a single JSON object to stdout:

```json
{"text": "display text", "color": "cyan"}
```

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | The text to display in the status line. Keep it short (under ~30 chars). |
| `color` | string | Color name. Supported: `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `dim`, `gray`. Can also be empty for no color (user can set colors via config). |

**To show nothing** (e.g., when data is unavailable): write nothing to stdout and exit 0. The widget will be silently omitted from the status line.

**Errors:** Write error messages to stderr (not stdout). Exit with a non-zero code on failure. The engine will fall back to the cached result or omit the widget.

---

## Writing Plugins by Language

### Shell Script

The simplest approach. Great for wrapping existing CLI tools.

```sh
#!/bin/sh
cat > /dev/null

# Example: show CPU temperature on macOS
temp=$(sysctl -n machdep.xcpm.cpu_thermal_level 2>/dev/null)
if [ -z "$temp" ]; then
    exit 0  # No output = widget hidden
fi

printf '{"text": "🌡️ %s°", "color": "yellow"}' "$temp"
```

**`plugin.json`:**
```json
{
  "name": "cpu-temp",
  "version": "1.0.0",
  "description": "CPU thermal level",
  "entry": "widget.sh",
  "cache_ttl": "30s",
  "interpreter": "sh"
}
```

### Python

Use the ccstatuswidgets Python SDK for the simplest approach, or handle the protocol manually.

#### With SDK

```bash
pip install ccstatuswidgets
```

```python
from ccstatuswidgets import widget

@widget(name="todo-count")
def render(input_data, config):
    import subprocess
    result = subprocess.run(["grep", "-rc", "TODO", "."], capture_output=True, text=True)
    count = len(result.stdout.strip().split("\n")) if result.stdout.strip() else 0
    if count == 0:
        return None  # Hide widget
    return {"text": f"📝 {count} TODOs", "color": "yellow"}
```

**`plugin.json`:**
```json
{
  "name": "todo-count",
  "version": "1.0.0",
  "description": "Count TODOs in project",
  "entry": "widget.py",
  "cache_ttl": "1m",
  "interpreter": "python3"
}
```

#### Without SDK

```python
import json
import sys

def main():
    # Read stdin
    input_data = json.loads(sys.stdin.read())

    # Your logic
    project = input_data.get("workspace", {}).get("current_dir", "")

    # Write output
    json.dump({"text": f"📁 {project}", "color": "cyan"}, sys.stdout)

if __name__ == "__main__":
    main()
```

### Node.js

```javascript
const data = require('fs').readFileSync('/dev/stdin', 'utf8');
const input = JSON.parse(data);

const tokens = input.context_window?.total_input_tokens || 0;
const cost = input.cost?.total_cost_usd || 0;

if (cost === 0) process.exit(0); // Hide widget

const output = {
    text: `💰 $${cost.toFixed(2)}/hr`,
    color: "yellow"
};

process.stdout.write(JSON.stringify(output));
```

**`plugin.json`:**
```json
{
  "name": "cost-rate",
  "version": "1.0.0",
  "description": "Cost per hour estimate",
  "entry": "widget.js",
  "interpreter": "node"
}
```

### Go

For maximum performance, compile to a binary:

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strings"
)

func main() {
    io.ReadAll(os.Stdin) // consume stdin

    out, err := exec.Command("df", "-h", "/").Output()
    if err != nil {
        os.Exit(0)
    }

    lines := strings.Split(string(out), "\n")
    if len(lines) < 2 {
        os.Exit(0)
    }

    fields := strings.Fields(lines[1])
    if len(fields) < 5 {
        os.Exit(0)
    }

    result := map[string]string{
        "text":  fmt.Sprintf("💾 %s used", fields[4]),
        "color": "dim",
    }
    json.NewEncoder(os.Stdout).Encode(result)
}
```

Build and set entry to the binary:
```json
{
  "name": "disk-usage",
  "version": "1.0.0",
  "description": "Disk usage percentage",
  "entry": "disk-usage"
}
```

No `interpreter` needed — the binary is directly executable.

---

## Using Input Data

The `StatusLineInput` on stdin contains useful context about the current Claude Code session. Here are some ideas:

| Field | What You Can Do |
|-------|-----------------|
| `workspace.current_dir` | Show project-specific info (package.json version, git remote, etc.) |
| `context_window.used_percentage` | Alert when context is getting full |
| `cost.total_cost_usd` | Calculate cost per hour, cost per token |
| `rate_limits.five_hour.used_percentage` | Show burn rate, estimate remaining time |
| `model.display_name` | Conditionally show info based on model |
| `session_id` | Track per-session metrics |

### Accessing Widget Config

If users add config for your widget in `config.json`:

```json
"my-widget": { "api_key": "xxx", "city": "London" }
```

This config is **not** passed to your plugin via stdin (it's only available to built-in Go widgets). For plugin configuration, use environment variables or a separate config file in your plugin directory.

---

## Caching

The engine caches your plugin's output based on the `cache_ttl` in `plugin.json`. This means:

- Your plugin won't be called on every render — only when the cache expires
- If your plugin is slow or times out (500ms default), the cached result is shown
- First render after install may show nothing (cache is empty, plugin may timeout)

**Choose TTL based on how often your data changes:**

| Data Type | Recommended TTL |
|-----------|-----------------|
| Static/slow-changing (uptime, disk) | `5m` to `30m` |
| Moderate (weather, stocks) | `5m` to `10m` |
| Fast-changing (CPU, network) | `30s` to `1m` |
| Real-time (live scores) | `30s` to `2m` |

---

## Timeout

The engine gives each plugin 500ms to complete (configurable via `timeout_ms` in the user's config). If your plugin takes longer:

1. The engine uses the last cached result
2. Your plugin's process continues running in the background
3. On the next render, if your plugin completes in time, the new result is cached

**Tips for staying under 500ms:**
- Use short HTTP timeouts (2-3 seconds max)
- Cache expensive computations in a local file
- Avoid blocking operations
- For network calls, the first render may timeout — that's OK, the cache catches up

---

## Testing

### Manual testing

```bash
# Test with empty input
echo '{}' | sh widget.sh

# Test with realistic input
echo '{"model":{"display_name":"Opus 4.6"},"workspace":{"current_dir":"/tmp"},"context_window":{"used_percentage":50}}' | sh widget.sh

# Test with interpreter
echo '{}' | python3 widget.py

# Verify valid JSON output
echo '{}' | sh widget.sh | python3 -m json.tool
```

### Test locally without publishing

```bash
# Copy plugin to the plugins directory
cp -r . ~/.ccstatuswidgets/plugins/my-ccw-plugin/

# Enable it
ccw add my-widget

# Check it appears
ccw list

# See it in action
ccw preview
```

### Verify with ccw doctor

```bash
ccw doctor
# Should show your plugin as installed

ccw plugin list
# Should show name, version, description
```

---

## Publishing

### 1. Push to GitHub

```bash
cd my-ccw-plugin
git init
git add .
git commit -m "feat: initial release"
gh repo create yourname/ccw-plugin-mywidget --public --source=. --push
```

### 2. Naming convention

Use the prefix `ccw-plugin-` for discoverability:
- `ccw-plugin-weather-extended`
- `ccw-plugin-spotify`
- `ccw-plugin-kubernetes`

### 3. README template

Include in your plugin's README:

```markdown
# ccw-plugin-mywidget

Description of what it shows.

## Install

\```bash
ccw plugin add github.com/yourname/ccw-plugin-mywidget
ccw add my-widget
\```

## Configuration

Describe any config needed.

## Requirements

List any dependencies (Python 3, curl, specific CLI tools, etc.)

## License

MIT
```

### 4. Users install with

```bash
ccw plugin add github.com/yourname/ccw-plugin-mywidget
ccw add my-widget
```

---

## Plugin Ideas

Looking for inspiration? Here are some plugins waiting to be built:

| Idea | Description | Complexity |
|------|-------------|------------|
| `cpu-usage` | Current CPU usage percentage | Easy |
| `disk-usage` | Disk space used/free | Easy |
| `network-speed` | Current upload/download speed | Medium |
| `kubernetes` | Current k8s context and namespace | Easy |
| `aws-profile` | Active AWS profile name | Easy |
| `node-version` | Node.js version in current project | Easy |
| `python-venv` | Active Python virtualenv name | Easy |
| `spotify` | Currently playing on Spotify (API) | Medium |
| `github-notifications` | Unread GitHub notification count | Medium |
| `ci-status` | Last CI/CD pipeline status | Medium |
| `countdown` | Countdown to a configured date | Easy |
| `world-clock` | Time in another timezone | Easy |
| `crypto` | Cryptocurrency prices | Medium |
| `ssh-agent` | Number of loaded SSH keys | Easy |
| `tmux-sessions` | Active tmux session count | Easy |

---

## Troubleshooting

### Plugin not showing up

1. Check `ccw plugin list` — is it installed?
2. Check `ccw list` — is it enabled?
3. Test manually: `echo '{}' | sh ~/.ccstatuswidgets/plugins/your-plugin/widget.sh`
4. Check for errors: `echo '{}' | sh ~/.ccstatuswidgets/plugins/your-plugin/widget.sh 2>&1`
5. Make sure `plugin.json` has the correct `entry` and `interpreter`

### Plugin output is empty

- Your plugin might be timing out (>500ms). Check if it runs slowly.
- The first render after install may be empty — the cache hasn't been populated yet. Wait for the next render.
- If returning nothing intentionally, that's correct — the widget is hidden.

### Plugin shows stale data

- Check `cache_ttl` in `plugin.json`. Lower it for faster updates.
- Clear the cache: `rm ~/.ccstatuswidgets/cache/your-widget.json`

### JSON parse errors

- Make sure your plugin outputs **only** the JSON object to stdout
- Debug output, logging, and errors should go to stderr: `echo "debug info" >&2`
- Verify with: `echo '{}' | sh widget.sh | python3 -m json.tool`

### Permission denied

- Make sure the entry file is executable: `chmod +x widget.sh`
- Or use the `interpreter` field in `plugin.json` to avoid needing execute permission

---

## Examples

See the official plugins for reference implementations:

- [ccw-plugin-uptime](https://github.com/warunacds/ccw-plugin-uptime) — Shell script, system command
- [ccw-plugin-battery](https://github.com/warunacds/ccw-plugin-battery) — Shell script, cross-platform (macOS/Linux)
- [ccw-plugin-docker](https://github.com/warunacds/ccw-plugin-docker) — Shell script, CLI tool integration
- [ccw-plugin-ip](https://github.com/warunacds/ccw-plugin-ip) — Shell script, HTTP call with fallback
