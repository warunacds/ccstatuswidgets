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

## Configuration

Config lives at `~/.ccstatuswidgets/config.json`:

```json
{
  "timeout_ms": 500,
  "lines": [
    { "widgets": ["model", "effort", "directory", "git-branch", "context-bar", "usage-5h", "usage-7d"] },
    { "widgets": ["lines-changed", "cost", "memory"] }
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

Bar-based widgets (context-bar, usage-5h, usage-7d) change color based on percentage: green (<50%), yellow (50-79%), red (80%+).

## CLI commands

| Command | Description |
|---------|-------------|
| `ccw` | Pipeline mode. Reads JSON from stdin, outputs rendered status line. Called by Claude Code automatically. |
| `ccw init` | Creates config directory, writes default config, patches Claude Code settings. |
| `ccw doctor` | Checks installation health: config validity, Claude Code settings, git and python3 availability. |
| `ccw preview` | Renders a sample status line using realistic mock data. Useful for testing your config. |
| `ccw version` | Prints version information. |

## Useful commands

```bash
# Test your config without starting Claude Code
ccw preview

# Diagnose setup issues
ccw doctor
```

## License

MIT -- see [LICENSE](LICENSE).
