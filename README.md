<div align="center">

# ⚡ ccstatuswidgets

### A customizable, plugin-ready status line for Claude Code CLI

`model` · `tokens` · `git` · `weather` · `stocks` · `cricket` · `flights` · `pomodoro` · `and more`

**23 built-in widgets** · **plugin system** · **single Go binary**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/warunacds/ccstatuswidgets/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/warunacds/ccstatuswidgets)](https://github.com/warunacds/ccstatuswidgets/releases)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/warunacds/ccstatuswidgets/graphs/commit-activity)

![Demo](https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/screenshots/demo.png)

</div>

---

## 📚 Table of Contents

- [Features](#-features)
- [Quick Start](#-quick-start)
- [Built-in Widgets](#-built-in-widgets)
- [Configuration](#-configuration)
- [Managing Widgets](#-managing-widgets)
- [CLI Commands](#-cli-commands)
- [Plugin System](#-plugin-system)
- [Creating Plugins](#-creating-plugins)
- [Development](#-development)
- [License](#-license)

---

## ✨ Features

- **📊 23 Built-in Widgets** — model, tokens, git status, weather, stocks, cricket scores, moon phase, and more
- **⚡ Fast** — Go binary with concurrent widget execution, renders in <10ms for built-in widgets
- **🔌 Plugin System** — extend with community plugins via simple shell/Python scripts
- **🎯 CLI Management** — `ccw add`, `ccw remove`, `ccw list` to manage widgets without editing JSON
- **📐 Multi-line Layout** — arrange widgets across multiple status lines, any order
- **🔄 Cache Fallback** — slow widgets (HTTP/git) fall back to cached results, per-widget TTL
- **🍅 Pomodoro Timer** — built-in work/break timer with `ccw pomo` commands
- **🌍 Cross-platform** — macOS and Linux, no external dependencies

---

## 🚀 Quick Start

### Install

```bash
curl -sSL https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/install.sh | sh
ccw init
```

That's it. Start Claude Code and the status line appears automatically.

### What `ccw init` does

1. Creates `~/.ccstatuswidgets/` with default config
2. Patches `~/.claude/settings.json` to point `statusLine.command` to `ccw`
3. Sets up cache and plugin directories

### Updating

Re-run the install script — your config is preserved, only the binary is replaced:

```bash
curl -sSL https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/install.sh | sh
```

<details>
<summary><b>Install from source</b></summary>

```bash
git clone https://github.com/warunacds/ccstatuswidgets.git
cd ccstatuswidgets
go build -ldflags "-s -w -X github.com/warunacds/ccstatuswidgets/internal/cli.Version=$(git describe --tags)" -o ccw ./cmd/ccw
sudo cp ccw /usr/local/bin/ccw
ccw init
```

</details>

---

## 📊 Built-in Widgets

### Core Widgets

| Widget | Output | Description |
|--------|--------|-------------|
| `model` | `Opus 4.6` | Current model display name |
| `effort` | `(high)` | Effort level from Claude settings |
| `directory` | `myproject` | Basename of working directory |
| `git-branch` | `(main)` | Current git branch |
| `git-status` | `✓` or `✎ 3M 2U` | Working tree status (modified/untracked/added/deleted) |
| `git-diff` | `staged: 4 files` | Staged changes count |
| `tokens` | `↑12.4k ↓8.2k` | Input/output token counts |
| `total-tokens` | `20.6k tokens` | Combined token count |
| `session-time` | `⏱ 1h23m` | Elapsed session time |
| `context-bar` | `ctx ████░░░░░░ 35%` | Context window usage bar |
| `usage-5h` | `5h ██░░░░░░░░ 22% -8% ↻2h30m` | 5-hour rate limit with pace tracking |
| `usage-7d` | `7d █░░░░░░░░░ 8% -21% ↻3d1h` | 7-day rate limit with pace tracking |
| `lines-changed` | `+342 -87` | Lines added (green) / removed (red) |
| `cost` | `api eq. $1.47` | Session cost (detects Max plan) |
| `memory` | `542MB` | Parent process memory |

### Fun Widgets

| Widget | Output | Description | Needs API Key? |
|--------|--------|-------------|----------------|
| `weather` | `🌤 +33°C` | Current weather via [wttr.in](https://wttr.in) | No |
| `now-playing` | `♪ Song - Artist` | Currently playing track (macOS + Linux) | No |
| `hackernews` | `HN: Story Title` | Top Hacker News story | No |
| `moon` | `🌒 waxing crescent` | Current moon phase | No |
| `stocks` | `AAPL +1.2% TSLA -3.1%` | Stock price changes | No |
| `pomodoro` | `🍅 18:42` / `☕ 3:21` | Pomodoro work/break timer | No |
| `flight` | `✈ UL504 ⬆ active` | Live flight tracking | Yes (AviationStack) |
| `cricket` | `🏏 SL 245/3 (42.1)` | Live cricket scores via ESPN | No |

> Bar widgets change color based on usage: 🟢 green (<50%), 🟡 yellow (50–79%), 🔴 red (80%+)

---

## ⚙️ Configuration

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
    "weather":     { "city": "Colombo", "units": "metric" },
    "stocks":      { "symbols": ["AAPL", "TSLA"] },
    "pomodoro":    { "work_mins": 25, "break_mins": 5 }
  }
}
```

| Option | Description |
|--------|-------------|
| `timeout_ms` | Max time (ms) to wait for widgets before cache fallback. Default: `500` |
| `lines` | Array of line configs. Each line lists widget names in display order |
| `widgets` | Per-widget settings. Keys match widget names |

### Widget-specific config

<details>
<summary><b>Weather</b></summary>

```json
"weather": { "city": "Colombo", "units": "metric" }
```
- `city` — city name (default: IP geolocation)
- `units` — `"metric"` or `"imperial"`
</details>

<details>
<summary><b>Stocks</b></summary>

```json
"stocks": { "symbols": ["AAPL", "TSLA", "GOOG"] }
```
- `symbols` — array of stock ticker symbols
</details>

<details>
<summary><b>Flight</b></summary>

Supports two providers — choose one:

**Option 1: AviationStack** (default) — [aviationstack.com](https://aviationstack.com/), 500 requests/month free
```json
"flight": { "api_key": "your_key", "flight": "UL504" }
```

**Option 2: AeroDataBox** via RapidAPI — [rapidapi.com/aedbx](https://rapidapi.com/aedbx-aedbx/api/aerodatabox), more generous free tier
```json
"flight": { "api_key": "your_rapidapi_key", "flight": "UL504", "provider": "aerodatabox" }
```

- `api_key` — API key for your chosen provider
- `flight` — IATA flight number (e.g., `UL504`, `QF1`, `BA256`)
- `provider` — `"aviationstack"` (default) or `"aerodatabox"`

You can also set the flight dynamically: `ccw track UL504`
</details>

<details>
<summary><b>Cricket</b></summary>

Uses ESPN's free cricket API — **no API key required**.

```json
"cricket": { "team": "SL" }
```
- `team` — optional filter to show only matches for your team (e.g., `SL`, `AUS`, `IND`, `RCB`)
- `leagues` — optional array of ESPN league IDs to check (default: `["8048", "8676"]` for IPL + International)

Common league IDs: `8048` (IPL), `8676` (International), `8044` (Big Bash), `8052` (CPL)
</details>

<details>
<summary><b>Pomodoro</b></summary>

```json
"pomodoro": { "work_mins": 25, "break_mins": 5 }
```
</details>

<details>
<summary><b>Cache TTL</b></summary>

Any widget can have a custom cache TTL:
```json
"weather": { "city": "Colombo", "cache_ttl": "30m" }
```
Default: `5m`. Accepts Go duration strings (`30s`, `5m`, `2h`, `24h`).
</details>

---

## 🎯 Managing Widgets

No need to edit JSON — use the interactive configurator or CLI commands:

```bash
# Interactive mode — visual layout editor
ccw configure
```

```bash
# See all available widgets and their status
ccw list

# Enable a widget (appends to last row)
ccw add weather

# Enable at a specific position (row 2, column 4)
ccw add weather -r 2 -c 4
ccw add weather --rc 2:4      # shorthand
ccw add weather --rc=2:4      # also works

# Disable a widget
ccw remove weather

# Edit config manually
ccw config edit
```

### Example: 3-line setup

```bash
ccw add weather --rc 2:1
ccw add stocks --rc 3:1
ccw add hackernews --rc 3:2
ccw add moon --rc 3:3
```

---

## 🖥️ CLI Commands

### Core

| Command | Description |
|---------|-------------|
| `ccw` | Pipeline mode — reads JSON from stdin, outputs status line. Called by Claude Code. |
| `ccw init` | Creates config, patches Claude Code settings |
| `ccw doctor` | Checks installation health |
| `ccw preview` | Renders sample status line with mock data |
| `ccw version` | Prints version |

### Widget Management

| Command | Description |
|---------|-------------|
| `ccw list` | Show all widgets with enabled/disabled status |
| `ccw add <widget>` | Enable a widget |
| `ccw add <widget> --rc 2:4` | Enable at row:col position |
| `ccw remove <widget>` | Disable a widget |
| `ccw configure` | Interactive widget layout configurator |
| `ccw config edit` | Open config in `$EDITOR` |

### Shortcuts

| Command | Description |
|---------|-------------|
| `ccw pomo start` | Start a 25-minute work session |
| `ccw pomo stop` | Stop the timer |
| `ccw pomo skip` | Skip to next phase (work/break) |
| `ccw pomo status` | Show timer status |
| `ccw track UL504` | Set flight to track |
| `ccw track stop` | Stop tracking |
| `ccw hn` | Top 5 Hacker News stories in terminal |

---

## 🔌 Plugin System

Extend ccstatuswidgets with community plugins. Plugins are standalone executables that receive JSON on stdin and write JSON to stdout.

![Status line with plugins](https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/screenshots/demo-with-plugins.png)
*All built-in widgets + battery, uptime, docker plugins + IPL cricket scores*


### Official Plugins

| Plugin | Description | Install |
|--------|-------------|---------|
| [uptime](https://github.com/warunacds/ccw-plugin-uptime) | System uptime (`⬆ 3d 12h`) | `ccw plugin add github.com/warunacds/ccw-plugin-uptime` |
| [battery](https://github.com/warunacds/ccw-plugin-battery) | Battery level (`🔋 78%` / `⚡ 100%`) | `ccw plugin add github.com/warunacds/ccw-plugin-battery` |
| [docker](https://github.com/warunacds/ccw-plugin-docker) | Running containers (`🐳 4 containers`) | `ccw plugin add github.com/warunacds/ccw-plugin-docker` |
| [ip](https://github.com/warunacds/ccw-plugin-ip) | Public IP (`🌐 203.0.113.42`) | `ccw plugin add github.com/warunacds/ccw-plugin-ip` |

After installing, enable with `ccw add <name>`:

```bash
ccw plugin add github.com/warunacds/ccw-plugin-battery
ccw add battery                    # enable on the status line
ccw add battery --rc 2:5           # or at a specific position
```

### Managing Plugins

```bash
ccw plugin add github.com/user/my-plugin    # install
ccw plugin list                              # list installed
ccw plugin remove my-plugin                  # uninstall
ccw plugin update my-plugin                  # update
ccw plugin update --all                      # update all
```

---

## 🛠️ Creating Plugins

A plugin is a Git repo with a `plugin.json` and an executable entry point.

### Plugin Manifest

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

### Shell Plugin Example

```sh
#!/bin/sh
cat > /dev/null  # consume stdin
printf '{"text": "hello world", "color": "cyan"}'
```

### Python Plugin Example

Install the Python SDK:

```bash
pip install ccstatuswidgets
```

```python
from ccstatuswidgets import widget

@widget(name="my-widget")
def render(input_data, config):
    return {"text": "hello", "color": "green"}
```

### Plugin Protocol

Plugins receive `StatusLineInput` JSON on stdin and must write a `WidgetOutput` JSON object to stdout:

```json
{"text": "display text", "color": "cyan"}
```

Supported colors: `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `dim`, `gray`

---

## 🗺️ Roadmap

### Phase 3 (next)
- **🎨 Themes** — built-in themes (dark, light, solarized, dracula, nord) with `ccw theme` selector
- **🌈 Custom Colors** — per-widget foreground/background colors via config (`"color": "#ff6b6b"`, hex + 256-color support)
- **⚡ Powerline Mode** — arrow separators between widgets with Nerd Font support
- **🔤 Custom Separators** — configurable widget separators (pipes, dots, arrows, custom Unicode)
- **🖼️ Widget Formatting** — padding, bold, italic, background colors per widget
- **📐 Alignment** — left/center/right alignment with flex separators

### Future
- Windows support
- Homebrew tap (`brew install ccstatuswidgets`)
- Plugin registry (searchable catalog)
- More community plugins

---

## 🏗️ Development

### Prerequisites

- [Go](https://go.dev) 1.21+
- Git

### Setup

```bash
git clone https://github.com/warunacds/ccstatuswidgets.git
cd ccstatuswidgets
go test ./...        # run tests
go build -o ccw ./cmd/ccw   # build
```

### Project Structure

```
ccstatuswidgets/
├── cmd/ccw/                  # Entry point, CLI routing
├── internal/
│   ├── protocol/             # StatusLineInput, WidgetOutput types
│   ├── widget/               # Widget interface + registry
│   ├── widgets/              # 23 built-in widget implementations
│   ├── engine/               # Concurrent executor with cache fallback
│   ├── renderer/             # ANSI color renderer
│   ├── cache/                # File-based TTL cache
│   ├── config/               # JSON config loader + defaults
│   ├── httpclient/           # Shared HTTP client for API widgets
│   ├── plugin/               # External plugin runner
│   └── cli/                  # CLI commands (init, doctor, pomo, etc.)
├── python-sdk/               # Python SDK for plugin authors
├── install.sh                # Curl-based installer
├── .goreleaser.yml           # Cross-platform release config
└── .github/workflows/        # CI/CD — auto-release on tag
```

### Running Tests

```bash
go test ./...              # all tests
go test ./... -race        # with race detector
go vet ./...               # lint
```

### Releasing

Push a version tag — GitHub Actions builds binaries and creates the release automatically:

```bash
git tag v0.4.0
git push origin v0.4.0
```

---

## 📄 License

[MIT](LICENSE) © Waruna De Silva

---

<div align="center">

### 🌟 Show Your Support

Give a ⭐ if this project helped you!

[![GitHub stars](https://img.shields.io/github/stars/warunacds/ccstatuswidgets?style=social)](https://github.com/warunacds/ccstatuswidgets/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/warunacds/ccstatuswidgets?style=social)](https://github.com/warunacds/ccstatuswidgets/network/members)

[Report Bug](https://github.com/warunacds/ccstatuswidgets/issues) · [Request Feature](https://github.com/warunacds/ccstatuswidgets/issues)

</div>
