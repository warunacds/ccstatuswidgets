# ccstatuswidgets

## What This Is

A Go binary (`ccw`) that provides a customizable, multi-line status line for Claude Code CLI. Reads JSON from stdin, runs widgets concurrently, renders colored ANSI output to stdout.

**Repo:** `github.com/warunacds/ccstatuswidgets`
**Module:** `github.com/warunacds/ccstatuswidgets`
**Binary:** `ccw`
**Go version:** 1.21+
**Dependencies:** stdlib only (no external deps)
**License:** MIT

## Build & Test

```bash
go test ./...              # all tests
go test ./... -race        # with race detector
go vet ./...               # lint
go build -o ccw ./cmd/ccw  # build binary
```

## Architecture

```
stdin JSON (from Claude Code)
    │
    ▼
┌─────────────────┐
│  cmd/ccw/main.go │ ─── CLI routing (init, doctor, configure, pomo, etc.)
└────────┬────────┘
         │ pipeline mode (no args)
         ▼
┌─────────────────┐     ┌──────────────────┐
│  config.Load()  │ ──▸ │  widget.Registry │ ◂── widgets.RegisterAll() + plugin.DiscoverPlugins()
└────────┬────────┘     └────────┬─────────┘
         │                       │
         ▼                       ▼
┌─────────────────────────────────────┐
│           engine.Run()              │
│  Launches all widgets concurrently  │
│  500ms timeout per widget           │
│  Cache fallback on timeout/error    │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│         renderer.Render()           │
│  Per-widget styling (fg/bg/bold)    │
│  Custom separators / Powerline      │
│  ANSI color output                  │
└────────────────┬────────────────────┘
                 │
                 ▼
            stdout (colored text)
```

## Package Map

| Package | Responsibility | Key Types/Functions |
|---------|---------------|---------------------|
| `cmd/ccw` | Entry point, CLI routing | `main()`, `runPipeline()` |
| `internal/protocol` | Data types | `StatusLineInput`, `WidgetOutput`, `ModelInfo`, `ContextInfo`, `RateLimits`, `CostInfo` |
| `internal/widget` | Widget interface + registry | `Widget` interface (`Name()`, `Render()`), `Registry` (`Register`, `Get`, `Names`) |
| `internal/widgets` | 23 built-in widget implementations | One file per widget, `RegisterAll()` in `register.go` |
| `internal/engine` | Concurrent executor | `Engine.Run()` — goroutines per widget, timeout, cache fallback |
| `internal/renderer` | ANSI output | `Render()`, `WidgetStyle`, `FgCode`/`BgCode`, Powerline mode |
| `internal/cache` | File-based TTL cache | `Cache.Get()`, `Cache.Set()` — atomic writes in `~/.ccstatuswidgets/cache/` |
| `internal/config` | JSON config | `Config` struct, `Load()`, `Save()`, `Default()`, `ConfigDir()` |
| `internal/httpclient` | Shared HTTP client | `Client.Get()`, `Client.GetWithHeaders()` — 3s timeout |
| `internal/plugin` | External plugin runner | `DiscoverPlugins()`, `ExternalWidget` (implements `Widget` interface) |
| `internal/cli` | CLI commands | One file per command (init, doctor, configure, pomo, track, hn, plugin, list, add, remove, etc.) |

## Widget Interface

```go
type Widget interface {
    Name() string
    Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error)
}
```

- Return `nil, nil` to hide the widget (no data available)
- Return `nil, error` for errors (engine falls back to cache)
- `cfg` is the per-widget config from `config.json` `widgets` map
- For testability, widgets accept config overrides (e.g., `cfg["base_url"]`, `cfg["state_dir"]`)

## All 23 Built-in Widgets

| Widget | File | Source | Notes |
|--------|------|--------|-------|
| model | model.go | stdin | Display name, magenta |
| effort | effort.go | file | Reads ~/.claude/settings*.json |
| directory | directory.go | stdin | Basename of cwd, cyan |
| git-branch | gitbranch.go | git cmd | Branch name in parens, yellow |
| git-status | gitstatus.go | git cmd | ✓ clean or ✎ 3M 2U dirty |
| git-diff | gitdiff.go | git cmd | staged: N files |
| context-bar | context.go | stdin | Bar with %, uses bars.go helper |
| usage-5h | usage.go | stdin | Bar + pace tracking |
| usage-7d | usage.go | stdin | Bar + pace tracking |
| tokens | tokens.go | stdin | ↑12.4k ↓8.2k, uses formatTokenCount |
| total-tokens | totaltokens.go | stdin | Combined count, shares formatTokenCount |
| session-time | sessiontime.go | ps cmd | ⏱ 1h23m, reads parent process start |
| lines-changed | lines.go | stdin | Raw ANSI green/red |
| cost | cost.go | stdin | Detects Max plan via RateLimits != nil |
| memory | memory.go | ps cmd | RSS in MB, cfg["pid"] for testing |
| weather | weather.go | HTTP | wttr.in, no API key |
| now-playing | nowplaying.go | osascript/playerctl | Spotify + Music, macOS/Linux |
| hackernews | hackernews.go | HTTP | HN Firebase API |
| stocks | stocks.go | HTTP | Yahoo Finance, raw ANSI |
| flight | flight.go | HTTP | AviationStack or AeroDataBox, progress bar |
| cricket | cricket.go | HTTP | ESPN free API, no key needed |
| moon | moon.go | computation | Synodic month algorithm |
| pomodoro | pomodoro.go | file | Reads ~/.ccstatuswidgets/pomodoro.json |

## CLI Commands

| Command | Handler | File |
|---------|---------|------|
| (no args) | `runPipeline()` | main.go |
| init | `cli.RunInit()` | cli/init.go |
| reset | `cli.RunReset()` | cli/init.go |
| doctor | `cli.RunDoctor()` | cli/doctor.go |
| preview | `cli.RunPreview()` | cli/preview.go |
| version | `cli.RunVersion()` | cli/version.go |
| configure | `cli.RunConfigure()` | cli/configure.go |
| config edit | `cli.RunConfigEdit()` | cli/configedit.go |
| list | `cli.RunList()` | cli/list.go |
| add | `cli.RunAdd()` | cli/add.go |
| remove | `cli.RunRemove()` | cli/remove.go |
| plugin | `cli.RunPlugin()` | cli/plugin.go |
| pomo | `cli.RunPomo()` | cli/pomo.go |
| track | `cli.RunTrack()` | cli/track.go |
| hn | `cli.RunHN()` | cli/hn.go |

## Config

Lives at `~/.ccstatuswidgets/config.json`:

```json
{
  "timeout_ms": 500,
  "separator": " ",
  "powerline": false,
  "lines": [
    {"widgets": ["model", "effort", "directory", "git-branch", "context-bar", "tokens", "session-time", "usage-5h", "usage-7d"]},
    {"widgets": ["lines-changed", "git-status", "cost", "memory"]}
  ],
  "widgets": {
    "weather": {"city": "Colombo", "fg": "#f1fa8c"},
    "context-bar": {"bar_length": 10, "show_percentage": true}
  }
}
```

Per-widget styling keys: `fg`, `bg`, `bold`, `dim`, `italic`, `underline`, `cache_ttl`.

## Plugin System

Plugins are executables in `~/.ccstatuswidgets/plugins/<name>/` with a `plugin.json` manifest. They receive `StatusLineInput` JSON on stdin, write `WidgetOutput` JSON to stdout. See [PLUGINS.md](PLUGINS.md).

## Data Directory

```
~/.ccstatuswidgets/
├── config.json
├── cache/              # per-widget cache files (JSON with TTL)
├── plugins/            # external plugins (git repos)
└── pomodoro.json       # pomodoro timer state (created by ccw pomo start)
```

## Releasing

Push a tag — GitHub Actions builds and releases automatically:

```bash
git tag v0.7.0
git push origin v0.7.0
```

CI: `.github/workflows/release.yml` — builds 4 binaries (darwin/linux × amd64/arm64), creates GitHub release.

## Key Design Decisions

- **Stdlib only** — no external Go dependencies, keeps binary small and builds fast
- **Concurrent engine** — all widgets run in parallel with goroutines, 500ms timeout
- **Cache fallback** — slow/failing widgets show last cached result instead of blank
- **Per-widget cache TTL** — weather 30m, cricket 2m, etc. via `cache_ttl` config
- **nil,nil = hide** — widgets return nil to silently omit themselves (no data = no widget)
- **Atomic cache writes** — temp file + rename to prevent corruption
- **Raw ANSI passthrough** — widgets like stocks/lines-changed embed their own ANSI, renderer passes through
- **Plugin protocol** — same Widget interface, ExternalWidget adapter handles stdin/stdout piping
- **ccw init preserves config** — only writes defaults if config.json doesn't exist
- **Session time timezone** — `time.ParseInLocation` for ps output (local time, not UTC)
- **Flight providers** — supports both AviationStack and AeroDataBox via `provider` config key

## Common Patterns

### Adding a new built-in widget

1. Create `internal/widgets/mywidget.go` with struct implementing `Widget` interface
2. Create `internal/widgets/mywidget_test.go` with tests
3. Register in `internal/widgets/register.go` → `RegisterAll()`
4. For HTTP widgets: use `httpclient.New()`, accept `cfg["base_url"]` for testing
5. For command widgets: accept cfg overrides for testability

### Adding a new CLI command

1. Create `internal/cli/mycommand.go` with `RunMyCommand()` function
2. Create `internal/cli/mycommand_test.go`
3. Add case to switch in `cmd/ccw/main.go`
4. Update `printUsage()`
