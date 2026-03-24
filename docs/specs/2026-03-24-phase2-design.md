# ccstatuswidgets Phase 2 — Design Spec

## Overview

Phase 2 adds 8 new built-in widgets (weather, now-playing, flight, cricket, stocks, hackernews, moon, pomodoro) and an external plugin system for community extensibility. Built-in widgets are Go code for performance; the plugin system enables Python (or any language) plugins via a JSON stdin/stdout protocol.

## Goals

1. Ship 8 fun, useful widgets that work out of the box
2. Enable community plugin development without requiring Go
3. Keep first-party widgets fast (no Python runtime dependency)
4. Minimal changes to Phase 1 architecture

## Non-Goals

- TUI config editor
- Windows support
- Plugin auto-update
- Homebrew tap (deferred to Phase 3)

---

## Workstream 1: Built-in Widgets

### Widget Summary

| Widget | Source | API/Method | Config Keys | Cache TTL | Color |
|--------|--------|------------|-------------|-----------|-------|
| `weather` | HTTP | wttr.in (free, no key) | `city`, `units` (metric/imperial) | 30m | yellow |
| `now-playing` | macOS osascript | Queries active media app | (none) | none | magenta |
| `flight` | HTTP | AviationStack free tier | `api_key`, `flight` | 5m | cyan |
| `cricket` | HTTP | Free cricket API/scrape | `team` (optional filter) | 2m | green |
| `stocks` | HTTP | Yahoo Finance (no key needed) | `symbols` | 5m | green/red |
| `hackernews` | HTTP | HN API (free, no key) | `show_score` (bool) | 10m | yellow |
| `moon` | Computation | Astronomical calculation | (none) | 24h | dim |
| `pomodoro` | File-based | Reads `~/.ccstatuswidgets/pomodoro.json` | `work_mins`, `break_mins` | none | red/green |

### Output Formats

- weather: `☀ 28°C` or `🌧 15°C`
- now-playing: `♪ Song - Artist` (truncated to ~30 chars)
- flight: `✈ UL504 ⬆ cruising` or `✈ UL504 landing 14:30`
- cricket: `🏏 SL 245/3 (42.1)` or `🏏 SL v AUS - SL won`
- stocks: `AAPL +1.2%` (green if up, red if down)
- hackernews: `HN: Title of top story`
- moon: `🌗 waning gibbous`
- pomodoro: `🍅 18:42` (work, red) or `☕ 3:21` (break, green)

### Widget Details

#### Weather (`internal/widgets/weather.go`)

Uses wttr.in — a free weather service that requires no API key.

Request: `GET https://wttr.in/{city}?format=%c+%t` (returns emoji + temp).

Config:
```json
"weather": {
  "city": "Colombo",
  "units": "metric"
}
```

Falls back to IP-based geolocation if no city configured. Returns nil if HTTP fails (cache fallback handles stale data).

#### Now Playing (`internal/widgets/nowplaying.go`)

macOS only. Uses `osascript` to query the system media player — works with Spotify, Apple Music, YouTube in browser, and any app that reports to macOS media controls.

```applescript
tell application "System Events"
    set frontApp to name of first process whose frontmost is true
end tell
-- Then query the active media app for track/artist
```

Simpler approach: use `osascript` to read from the Now Playing notification center or query known media apps in order (Spotify → Music → generic).

Returns nil on non-macOS or if nothing is playing.

#### Flight (`internal/widgets/flight.go`)

Uses AviationStack free tier API (500 requests/month on free plan).

Config:
```json
"flight": {
  "api_key": "your_aviationstack_key",
  "flight": "UL504"
}
```

Request: `GET http://api.aviationstack.com/v1/flights?access_key={key}&flight_iata={flight}`

Parses flight status (scheduled, active, landed, cancelled), altitude, and ETA. Returns nil if no flight configured or API fails.

#### Cricket (`internal/widgets/cricket.go`)

Uses a free cricket scores API (e.g., cricketdata.org free tier or Cricbuzz unofficial API).

Config:
```json
"cricket": {
  "api_key": "your_key",
  "team": "SL"
}
```

Shows live match score if a match is in progress, or the most recent result. `team` filter is optional — if set, only shows matches involving that team. Returns nil if no live matches.

#### Stocks (`internal/widgets/stocks.go`)

Uses Yahoo Finance API (no key required) — fetches quote data for configured symbols.

Config:
```json
"stocks": {
  "symbols": ["AAPL", "TSLA"]
}
```

Shows each symbol with daily percentage change. Multiple symbols separated by spaces: `AAPL +1.2% TSLA -3.1%`. Green for positive, red for negative — uses raw ANSI (same pattern as lines-changed widget).

Returns nil if no symbols configured or market data unavailable.

#### Hacker News (`internal/widgets/hackernews.go`)

Uses the official HN API (free, no key): `https://hacker-news.firebaseio.com/v0/topstories.json` then fetches the top story.

Config:
```json
"hackernews": {
  "show_score": true
}
```

Fetches top story ID, then story details. Output: `HN: Story Title` or `HN: Story Title (342pts)` if show_score is true. Title truncated to ~40 chars.

#### Moon Phase (`internal/widgets/moon.go`)

Pure computation — no network calls. Uses the synodic month algorithm to calculate current moon phase from the current date.

The 8 phases: new moon, waxing crescent, first quarter, waxing gibbous, full moon, waning gibbous, last quarter, waning crescent.

Each phase gets an emoji: 🌑🌒🌓🌔🌕🌖🌗🌘

Output: `🌗 waning gibbous`

#### Pomodoro (`internal/widgets/pomodoro.go`)

Reads timer state from `~/.ccstatuswidgets/pomodoro.json`:

```json
{
  "phase": "work",
  "started_at": 1711300000,
  "work_mins": 25,
  "break_mins": 5
}
```

Computes remaining time from `started_at` + duration. When timer expires, auto-transitions (work→break→work). Shows countdown: `🍅 18:42` (work, red) or `☕ 3:21` (break, green).

Returns nil if no pomodoro.json exists (timer not started).

### Shared HTTP Client (`internal/httpclient/client.go`)

Thin wrapper around `net/http` for the 5 HTTP-calling widgets:

```go
type Client struct {
    httpClient *http.Client
    userAgent  string
}

func New() *Client  // 3-second timeout, ccw/<version> user-agent
func (c *Client) Get(url string) ([]byte, error)  // returns body bytes
```

3-second timeout is well above the 500ms widget timeout — on first run an HTTP widget will likely timeout and show nothing, then cache the result on the next successful run. Subsequent renders use cache fallback if the HTTP call is slow.

### API Key Configuration

All API keys and secrets live in `~/.ccstatuswidgets/config.json` under each widget's config section:

```json
{
  "widgets": {
    "weather": {"city": "Colombo", "units": "metric"},
    "flight": {"api_key": "xxx", "flight": "UL504"},
    "cricket": {"api_key": "xxx", "team": "SL"},
    "stocks": {"symbols": ["AAPL", "TSLA"]}
  }
}
```

Widgets that need an API key return nil if the key is not configured.

---

## Workstream 2: Plugin System

### Protocol

An external plugin is a standalone executable. The engine passes `StatusLineInput` JSON on stdin; the plugin writes `WidgetOutput` JSON to stdout. Same contract as built-in widgets.

```
Engine → stdin JSON → Plugin Executable → stdout JSON → Engine
```

Timeout: same as built-in widgets (default 500ms). Cache fallback applies identically.

### Plugin Directory Structure

```
~/.ccstatuswidgets/plugins/
└── my-widget/
    ├── plugin.json      # manifest
    └── main.py          # entry point (or any executable)
```

### Plugin Manifest (`plugin.json`)

```json
{
  "name": "my-widget",
  "version": "1.0.0",
  "description": "Does something cool",
  "entry": "main.py",
  "cache_ttl": "5m",
  "interpreter": "python3"
}
```

Fields:
- `name` — widget identifier (used in config.json `lines`)
- `version` — semver
- `description` — one-line description
- `entry` — relative path to the entry point file
- `cache_ttl` — cache duration for this plugin's output (default: 5m)
- `interpreter` — (optional) runtime to execute the entry file. If omitted, the entry file must be directly executable. Common values: `python3`, `node`, `ruby`.

### Plugin Runner (`internal/plugin/runner.go`)

At startup, the engine scans `~/.ccstatuswidgets/plugins/`, reads each `plugin.json`, and creates an `ExternalWidget` adapter:

```go
type ExternalWidget struct {
    name        string
    dir         string
    entry       string
    interpreter string
    cacheTTL    time.Duration
}

func (w *ExternalWidget) Name() string { return w.name }
func (w *ExternalWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error)
```

`Render` executes the plugin: pipes `StatusLineInput` JSON to stdin, reads `WidgetOutput` JSON from stdout. The `ExternalWidget` implements `widget.Widget`, so it registers in the same registry as built-in widgets. The engine doesn't know the difference.

`DiscoverPlugins(pluginsDir string) ([]*ExternalWidget, error)` scans the plugins directory, reads manifests, and returns all discovered plugins.

### Plugin CLI (`internal/cli/plugin.go`)

```
ccw plugin add <github-url>    Clone repo into plugins dir, validate plugin.json
ccw plugin list                List installed plugins (name, version, description)
ccw plugin remove <name>       Remove plugin directory
ccw plugin update <name>       Git pull latest in plugin directory
ccw plugin update --all        Update all plugins
```

**`ccw plugin add github.com/user/repo`:**
1. `git clone` the repo into `~/.ccstatuswidgets/plugins/<repo-name>/`
2. Validate `plugin.json` exists and is valid
3. Check for name conflicts with built-in widgets
4. Print success message: "Installed <name>. Add it to your config: ccw config edit"

**`ccw plugin remove <name>`:**
1. Find plugin directory by name
2. Remove the directory
3. Print confirmation

### Python SDK

Published as `pip install ccstatuswidgets`. Provides a decorator that handles the stdin/stdout JSON protocol:

```python
from ccstatuswidgets import widget

@widget(name="my-widget")
def render(input, config):
    # input is a dict with model, workspace, context_window, etc.
    # config is the widget's config from config.json
    return {"text": "hello", "color": "green"}
```

The SDK:
1. Reads stdin, parses JSON into a dict
2. Calls the decorated function
3. Writes the returned dict as JSON to stdout
4. Handles errors (prints to stderr, exits cleanly)

The SDK is a single Python file (~50 lines). Published to PyPI.

---

## Workstream 3: Shortcut Commands

### Pomodoro Commands (`internal/cli/pomo.go`)

```
ccw pomo start     Start a 25-min work timer
ccw pomo stop      Stop the timer (removes state file)
ccw pomo skip      Skip to next phase (work→break, break→work)
ccw pomo status    Print current timer state
```

All commands read/write `~/.ccstatuswidgets/pomodoro.json`. The pomodoro widget reads this file each render.

State file:
```json
{
  "phase": "work",
  "started_at": 1711300000,
  "work_mins": 25,
  "break_mins": 5
}
```

`start` creates the file. `stop` deletes it. `skip` updates `phase` and resets `started_at`. `status` prints remaining time and current phase.

Work/break durations are configurable via config.json:
```json
"pomodoro": {
  "work_mins": 25,
  "break_mins": 5
}
```

### Flight Tracking (`internal/cli/track.go`)

```
ccw track UL504    Set flight number in config
ccw track stop     Remove flight from config
```

`track` reads config.json, sets `"flight": {"flight": "UL504"}`, writes it back. `stop` removes the flight key.

### Hacker News (`internal/cli/hn.go`)

```
ccw hn             Print top 5 HN stories with links
```

Fetches top 5 story IDs from HN API, fetches each story's details, prints:
```
1. Story Title (342 pts)
   https://example.com/article

2. Another Story (218 pts)
   https://news.ycombinator.com/item?id=12345
```

This is a richer output than the widget (which shows only the top story in one line).

---

## Architecture Changes to Phase 1

### 1. Engine — Per-Widget Cache TTL

Current: `cacheTTL = 5 * time.Minute` hardcoded constant.

Change: Read `cache_ttl` from widget config map. Parse duration strings (`"30m"`, `"2m"`, `"24h"`). Default: 5 minutes if not specified.

```go
func (e *Engine) getCacheTTL(widgetCfg map[string]interface{}) time.Duration {
    if widgetCfg == nil {
        return defaultCacheTTL
    }
    if v, ok := widgetCfg["cache_ttl"]; ok {
        if s, ok := v.(string); ok {
            if d, err := time.ParseDuration(s); err == nil {
                return d
            }
        }
    }
    return defaultCacheTTL
}
```

### 2. Engine — External Plugin Registration

In `main.go`, after `widgets.RegisterAll(registry)`:

```go
externalWidgets, _ := plugin.DiscoverPlugins(filepath.Join(configDir, "plugins"))
for _, ew := range externalWidgets {
    registry.Register(ew)
}
```

External widgets are registered in the same registry. The engine treats them identically.

### 3. CLI Routing — New Commands

Add to main.go switch:
- `plugin` → `cli.RunPlugin(os.Args[2:])`
- `pomo` → `cli.RunPomo(os.Args[2:])`
- `track` → `cli.RunTrack(os.Args[2:])`
- `hn` → `cli.RunHN()`

### 4. Config Defaults — Unchanged

`ccw init` still writes Phase 1 defaults. Users opt into Phase 2 widgets by editing config.json. No surprise widgets appearing after upgrade.

### 5. Preview Command — Updated

`ccw preview` updated to include sample Phase 2 widget data on a third line, showing what the new widgets look like.

---

## New Files Summary

### Workstream 1 (Built-in Widgets)
```
internal/httpclient/client.go          # shared HTTP client
internal/httpclient/client_test.go
internal/widgets/weather.go            # weather widget
internal/widgets/weather_test.go
internal/widgets/nowplaying.go         # now-playing widget
internal/widgets/nowplaying_test.go
internal/widgets/flight.go             # flight tracking widget
internal/widgets/flight_test.go
internal/widgets/cricket.go            # cricket scores widget
internal/widgets/cricket_test.go
internal/widgets/stocks.go             # stock prices widget
internal/widgets/stocks_test.go
internal/widgets/hackernews.go         # hacker news widget
internal/widgets/hackernews_test.go
internal/widgets/moon.go               # moon phase widget
internal/widgets/moon_test.go
internal/widgets/pomodoro.go           # pomodoro timer widget
internal/widgets/pomodoro_test.go
```

### Workstream 2 (Plugin System)
```
internal/plugin/runner.go              # discovers + executes external plugins
internal/plugin/runner_test.go
internal/cli/plugin.go                 # ccw plugin add/list/remove/update
internal/cli/plugin_test.go
python-sdk/                            # Python SDK (separate publishable package)
  ccstatuswidgets/__init__.py
  ccstatuswidgets/sdk.py
  setup.py
  README.md
```

### Workstream 3 (Shortcut Commands)
```
internal/cli/pomo.go                   # ccw pomo start/stop/skip/status
internal/cli/pomo_test.go
internal/cli/track.go                  # ccw track <flight>/stop
internal/cli/track_test.go
internal/cli/hn.go                     # ccw hn (top 5 stories)
internal/cli/hn_test.go
```

### Modified Files
```
internal/engine/engine.go              # per-widget cache TTL
internal/widgets/register.go           # register 8 new widgets
internal/cli/preview.go                # updated sample data
cmd/ccw/main.go                        # new CLI routes + plugin discovery
```

---

## Testing Strategy

- **HTTP widgets:** Mock the HTTP client in tests. Don't hit real APIs.
- **Now-playing:** Test the osascript output parsing. Mock the command execution.
- **Moon:** Verify against known dates (e.g., 2026-01-01 was a waxing gibbous).
- **Pomodoro:** Test state file read/write, timer expiry, phase transitions.
- **Plugin runner:** Create a test plugin executable, verify stdin/stdout protocol.
- **Plugin CLI:** Test add/remove/list with temp directories.
- **Shortcut commands:** Test config modification and state file management.
