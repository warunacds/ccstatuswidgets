# ccstatuswidgets

Customizable, plugin-ready status line for Claude Code CLI.

## What it does

ccstatuswidgets reads Claude Code's status line JSON from stdin, runs configured widgets concurrently with a timeout, and renders colored ANSI output. Widgets are arranged across multiple lines, each producing a text segment with an optional color. Results are cached so that slow widgets (like git operations) fall back gracefully.

## Example output

```
Opus 4.6 (high) myproject (main) ctx [||||......] 35% 5h [||........] 22% +2% ~2h30m 7d [|.........] 8% -6% ~5d0h
+342 -87 api eq. $1.47 312MB
```

Line 1: model, effort level, directory, git branch, context window bar, 5-hour and 7-day usage bars with pace indicators.
Line 2: lines added/removed, session cost, parent process memory.

## Quick start

```bash
curl -sSL https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/install.sh | sh
ccw init
```

`ccw init` creates `~/.ccstatuswidgets/config.json`, sets up cache and plugin directories, and patches Claude Code's `~/.claude/settings.json` to point the `statusLine.command` to the `ccw` binary.

### Updating

To update to the latest version, re-run the install script:

```bash
curl -sSL https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/install.sh | sh
```

Your config (`~/.ccstatuswidgets/config.json`) is preserved — only the binary is replaced.

If you installed from source:

```bash
cd ccstatuswidgets
git pull
go build -ldflags "-s -w -X github.com/warunacds/ccstatuswidgets/internal/cli.Version=$(git describe --tags)" -o ccw ./cmd/ccw
sudo cp ccw /usr/local/bin/ccw
```

## Configuration

Config lives at `~/.ccstatuswidgets/config.json`:

```json
{
  "timeout_ms": 500,
  "lines": [
    { "widgets": ["model", "effort", "directory", "git-branch", "context-bar", "tokens", "session-time", "usage-5h", "usage-7d"] },
    { "widgets": ["lines-changed", "git-status", "cost", "memory"] }
  ],
  "widgets": {
    "context-bar": { "bar_length": 10, "show_percentage": true },
    "usage-5h":    { "bar_length": 10, "show_percentage": true, "show_pace": true },
    "usage-7d":    { "bar_length": 10, "show_percentage": true, "show_pace": true },
    "cost":        { "detect_max_plan": true }
  }
}
```

| Option | Description |
|--------|-------------|
| `timeout_ms` | Maximum time (ms) to wait for all widgets before falling back to cache. Default: `500`. |
| `lines` | Array of line configs. Each line lists widget names in display order. |
| `widgets` | Per-widget overrides. Keys are widget names, values are widget-specific options. |

### Phase 2 widget configuration

Add Phase 2 widgets to your `lines` array, then configure them under `widgets`:

```json
{
  "lines": [
    { "widgets": ["model", "effort", "directory", "git-branch", "context-bar", "usage-5h", "usage-7d"] },
    { "widgets": ["lines-changed", "cost", "memory", "weather", "moon", "pomodoro"] },
    { "widgets": ["stocks", "hackernews", "now-playing"] }
  ],
  "widgets": {
    "weather": { "city": "Colombo", "units": "metric" },
    "stocks": { "symbols": ["AAPL", "TSLA"] },
    "cricket": { "api_key": "your_key", "team": "SL" },
    "flight": { "api_key": "your_key", "flight": "UL504" },
    "pomodoro": { "work_mins": 25, "break_mins": 5 }
  }
}
```

Widgets that require an API key (`cricket`, `flight`) will silently return nothing if the key is not configured. The `weather` widget uses wttr.in which requires no key. The `stocks` widget uses Yahoo Finance's public endpoint.

## Built-in widgets

| Widget | Description | Source | Color |
|--------|-------------|--------|-------|
| `model` | Current model display name | stdin JSON | magenta |
| `effort` | Effort level from Claude settings | `~/.claude/settings.json` | dim |
| `directory` | Basename of working directory | stdin JSON / `os.Getwd` | cyan |
| `git-branch` | Current git branch | `git symbolic-ref` | yellow |
| `context-bar` | Context window usage bar | stdin JSON | green/yellow/red |
| `usage-5h` | 5-hour rate limit bar with pace | stdin JSON | green/yellow/red |
| `usage-7d` | 7-day rate limit bar with pace | stdin JSON | green/yellow/red |
| `lines-changed` | Lines added and removed | stdin JSON | green(+) / red(-) |
| `cost` | Session cost in USD | stdin JSON | dim |
| `memory` | Parent process RSS memory | `ps` | dim |
| `weather` | Current weather via wttr.in | HTTP | yellow |
| `now-playing` | Currently playing track (macOS/Linux) | OS media API | magenta |
| `flight` | Live flight tracking via AviationStack | HTTP | cyan |
| `cricket` | Live cricket scores | HTTP | green |
| `stocks` | Stock price changes (green/red) | HTTP | raw ANSI |
| `hackernews` | Top Hacker News story | HTTP | yellow |
| `moon` | Current moon phase | computed | dim |
| `pomodoro` | Pomodoro work/break timer | local state file | red/green |
| `tokens` | Input/output token counts (`↑12.4k ↓8.2k`) | stdin JSON | dim |
| `total-tokens` | Combined token count (`20.6k tokens`) | stdin JSON | dim |
| `session-time` | Elapsed session time (`⏱ 1h23m`) | parent process | dim |
| `git-status` | Working tree status (`✓` or `✎ 3M 2U`) | `git status` | green/yellow |
| `git-diff` | Staged changes count (`staged: 4 files`) | `git diff` | green |

Bar-based widgets (context-bar, usage-5h, usage-7d) change color based on percentage: green (<50%), yellow (50-79%), red (80%+).

## CLI commands

| Command | Description |
|---------|-------------|
| `ccw` | Pipeline mode. Reads JSON from stdin, outputs rendered status line. Called by Claude Code automatically. |
| `ccw init` | Creates config directory, writes default config, patches Claude Code settings. |
| `ccw doctor` | Checks installation health: config validity, Claude Code settings, git and python3 availability. |
| `ccw preview` | Renders a sample status line using realistic mock data. Useful for testing your config. |
| `ccw version` | Prints version information. |
| `ccw pomo start` | Start a Pomodoro work session (25 min default). |
| `ccw pomo stop` | Stop the current Pomodoro timer. |
| `ccw pomo skip` | Skip to the next phase (work to break, or break to work). |
| `ccw pomo status` | Show the current Pomodoro timer status. |
| `ccw track <flight>` | Show live tracking info for a flight (e.g. `ccw track UL504`). |
| `ccw hn` | Display the top 5 Hacker News stories in your terminal. |
| `ccw list` | Show all available widgets with enabled/disabled status and position. |
| `ccw add <widget>` | Enable a widget. Appends to the last row by default. |
| `ccw add <widget> -r 2 -c 4` | Enable a widget at a specific row and column position. |
| `ccw add <widget> --rc 2:4` | Shorthand for row:col positioning. Also accepts `--rc=2:4`. |
| `ccw remove <widget>` | Disable a widget (removes it from the config layout). |
| `ccw config edit` | Open `config.json` in your `$EDITOR`. |
| `ccw plugin add <repo>` | Install a plugin from a Git repository. |
| `ccw plugin list` | List all installed plugins. |
| `ccw plugin remove <name>` | Remove an installed plugin. |

## Shortcut commands

```bash
# Pomodoro timer
ccw pomo start          # Start a 25-minute work session
ccw pomo skip           # Skip to break (or back to work)
ccw pomo status         # Check remaining time
ccw pomo stop           # Cancel the timer

# Flight tracking
ccw track UL504         # Live status for a flight

# Hacker News
ccw hn                  # Top 5 stories from HN
```

## Plugin system

ccstatuswidgets supports third-party plugins that add custom widgets to your status line.

### Installing plugins

```bash
ccw plugin add github.com/user/my-ccw-plugin
ccw plugin list
ccw plugin remove my-ccw-plugin
```

A plugin repository must contain a `plugin.json` at its root:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Shows my custom data",
  "command": "python3 main.py"
}
```

The plugin command is executed by the engine during each render cycle. It must print a JSON object to stdout:

```json
{"text": "my data", "color": "cyan"}
```

### Python SDK

For Python-based plugins, install the helper SDK:

```bash
pip install ccstatuswidgets
```

Example plugin using the SDK:

```python
from ccstatuswidgets import widget, output

@widget("my-widget")
def run(config):
    return output("Hello from plugin", color="cyan")
```

## Managing widgets

```bash
# See all available widgets and their status
ccw list

# Enable a widget (appends to last row)
ccw add weather

# Enable a widget at a specific position (row 2, column 4)
ccw add weather -r 2 -c 4
ccw add weather --rc 2:4    # shorthand
ccw add weather --rc=2:4   # also works

# Disable a widget
ccw remove weather

# Edit config manually
ccw config edit
```

## Useful commands

```bash
# Test your config without starting Claude Code
ccw preview

# Diagnose setup issues
ccw doctor
```

## License

MIT -- see [LICENSE](LICENSE).
