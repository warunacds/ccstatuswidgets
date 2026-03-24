# Phase 3: Styling & Formatting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add per-widget colors (hex, 256, named), text formatting (bold/dim/italic/underline), configurable separators, and Powerline mode to ccstatuswidgets.

**Architecture:** Extend the renderer with a style system that builds ANSI escape sequences from per-widget config. Colors are parsed from named, 256-color, or hex formats. The config schema gets new top-level fields (separator, powerline) and per-widget fields (fg, bg, bold, dim, italic, underline). Backward compatible — existing configs work unchanged.

**Tech Stack:** Go 1.21+, stdlib only

**Spec:** `docs/specs/2026-03-24-phase3-styling-design.md`

---

### Task 1: Color Parser

**Files:**
- Create: `internal/renderer/colors.go`
- Create: `internal/renderer/colors_test.go`

- [ ] **Step 1: Write color parser tests**

Test cases:
- Named color "red" → fg code "31", bg code "41"
- Named color "dim" → fg code "2"
- 256-color "196" → fg code "38;5;196", bg code "48;5;196"
- 256-color "color:42" → fg code "38;5;42"
- Hex "#ff6b6b" → fg code "38;2;255;107;107", bg code "48;2;255;107;107"
- Hex "#000000" → fg code "38;2;0;0;0"
- Empty string → empty code
- Invalid string "notacolor" → empty code
- All 9 named colors produce valid codes

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement color parser**

```go
func FgCode(color string) string   // returns ANSI fg code portion (no ESC prefix)
func BgCode(color string) string   // returns ANSI bg code portion
```

Named color map (fg/bg):
```
red: 31/41, green: 32/42, yellow: 33/43, blue: 34/44,
magenta: 35/45, cyan: 36/46, white: 37/47, dim: 2/-, gray: 90/-
```

256-color: parse int, return `38;5;N` / `48;5;N`
Hex: parse R,G,B from `#RRGGBB`, return `38;2;R;G;B` / `48;2;R;G;B`

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Commit**

```bash
git add internal/renderer/colors.go internal/renderer/colors_test.go
git commit -m "feat: color parser — named, 256-color, and hex truecolor support"
```

---

### Task 2: Style Builder

**Files:**
- Create: `internal/renderer/style.go`
- Create: `internal/renderer/style_test.go`

- [ ] **Step 1: Write style builder tests**

Test cases:
- Fg only → `\033[31m`
- Bg only → `\033[41m`
- Fg + bg → `\033[31;41m`
- Bold → `\033[1m`
- Fg + bold + italic → `\033[1;3;31m`
- All formatting: bold + dim + italic + underline + fg + bg
- No styles → empty string
- Hex fg + named bg
- Reset code is always `\033[0m`

The `WidgetStyle` struct:

```go
type WidgetStyle struct {
    Fg        string
    Bg        string
    Bold      bool
    Dim       bool
    Italic    bool
    Underline bool
}

func (s WidgetStyle) ANSI() string   // returns full ANSI prefix or empty
func (s WidgetStyle) Reset() string  // returns "\033[0m" or empty
```

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement style builder**

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Commit**

```bash
git add internal/renderer/style.go internal/renderer/style_test.go
git commit -m "feat: style builder — combine fg, bg, bold, dim, italic, underline into ANSI"
```

---

### Task 3: Config Schema Update

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/defaults.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: Add new fields to Config struct**

```go
type Config struct {
    TimeoutMs  int                                `json:"timeout_ms"`
    Separator  string                             `json:"separator"`
    Powerline  bool                               `json:"powerline"`
    Lines      []LineConfig                       `json:"lines"`
    Widgets    map[string]map[string]interface{}  `json:"widgets"`
}
```

- [ ] **Step 2: Update defaults**

Default separator: `" "` (space — preserves current behavior)
Default powerline: `false`

- [ ] **Step 3: Write tests**

Test: load config with separator and powerline fields, verify they round-trip correctly. Test default values.

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: config schema — separator and powerline fields"
```

---

### Task 4: Renderer — Styled Widget Output

**Files:**
- Modify: `internal/renderer/renderer.go`
- Modify: `internal/renderer/renderer_test.go`

- [ ] **Step 1: Write tests for styled rendering**

Test cases:
- Widget with `fg` in config renders with custom foreground color
- Widget with `fg` and `bg` renders both
- Widget with `bold: true` renders bold
- Widget with multiple formatting options
- Widget with no custom styling uses default `WidgetOutput.Color` (backward compat)
- Custom separator joins widgets
- Empty separator works

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Update renderer**

Change `Render` signature to accept config:
```go
func Render(lines [][]WidgetResult, cfg *config.Config) string
```

For each widget:
1. Read `fg`, `bg`, `bold`, `dim`, `italic`, `underline` from `cfg.Widgets[name]`
2. If no custom `fg`, fall back to `WidgetOutput.Color`
3. Build `WidgetStyle`, get ANSI prefix
4. Wrap text: `prefix + text + reset`
5. Join with `cfg.Separator` instead of hardcoded space

Update all callers of `Render` (main.go, preview.go, etc.) to pass config.

- [ ] **Step 4: Run ALL tests (not just renderer)**

The signature change affects callers. Update them.

- [ ] **Step 5: Commit**

```bash
git add internal/renderer/ cmd/ccw/main.go internal/cli/preview.go
git commit -m "feat: renderer supports per-widget colors, formatting, and custom separators"
```

---

### Task 5: Powerline Mode

**Files:**
- Modify: `internal/renderer/renderer.go`
- Modify: `internal/renderer/renderer_test.go`

- [ ] **Step 1: Write Powerline tests**

Test cases:
- Powerline enabled: widgets joined with arrow separator `` (U+E0B0)
- Arrow fg = previous widget bg, arrow bg = next widget bg
- First widget has left cap
- Last widget has right cap
- Widgets without bg get auto-assigned from default palette
- Powerline disabled: no arrow separators (regular separator)

- [ ] **Step 2: Run tests to verify they fail**

- [ ] **Step 3: Implement Powerline rendering**

Default palette for auto-assign: `["#44475a", "#6272a4", "#bd93f9", "#50fa7b", "#ffb86c", "#ff79c6", "#8be9fd", "#f1fa8c"]`

Powerline arrows:
- Right separator: `` (U+E0B0)
- Left separator: `` (U+E0B2)

For each widget pair, the arrow's fg is the left widget's bg, and the arrow's bg is the right widget's bg. This creates the seamless arrow effect.

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Commit**

```bash
git add internal/renderer/
git commit -m "feat: Powerline mode with arrow separators"
```

---

### Task 6: Interactive Color Picker in Configurator

**Files:**
- Modify: `internal/cli/configure.go`

- [ ] **Step 1: Add color editing (c key)**

When cursor is on a widget:
- Press `c` → show color picker menu
- Named colors listed with numbers (1-9)
- Option to type hex (#ff6b6b) or 256-color number
- Option 0 to clear custom color
- Save to `cfg.Widgets[name]["fg"]`

- [ ] **Step 2: Add background color editing (b key)**

Same as foreground but saves to `cfg.Widgets[name]["bg"]`

- [ ] **Step 3: Add formatting toggle (f key)**

Show current formatting state:
```
[x] bold  [ ] dim  [x] italic  [ ] underline
```
Press b/d/i/u to toggle, Enter to confirm.

- [ ] **Step 4: Add separator editing (/ key)**

Show current separator, let user type a new one.

- [ ] **Step 5: Add powerline toggle (p key)**

Toggle `cfg.Powerline` on/off.

- [ ] **Step 6: Update controls display**

Show new keys in the controls section.

- [ ] **Step 7: Build and test manually**

```bash
go build -o ccw ./cmd/ccw
./ccw configure
```

- [ ] **Step 8: Commit**

```bash
git add internal/cli/configure.go
git commit -m "feat: color picker, formatting, separator editing in configurator"
```

---

### Task 7: Wire and Polish

**Files:**
- Modify: `cmd/ccw/main.go`
- Modify: `internal/cli/preview.go`

- [ ] **Step 1: Update preview to show styled output**

Add some sample colors to the preview's config to demonstrate the styling.

- [ ] **Step 2: Final test run**

```bash
go test ./... -v -race
go vet ./...
go build -o ccw ./cmd/ccw
```

- [ ] **Step 3: Update README with styling examples**

Add a "Styling" section showing per-widget color config, separator options, and Powerline setup.

- [ ] **Step 4: Commit, push, tag**

```bash
git add .
git commit -m "feat: Phase 3 complete — styling, colors, Powerline"
git push
git tag v0.6.0
git push origin v0.6.0
```
