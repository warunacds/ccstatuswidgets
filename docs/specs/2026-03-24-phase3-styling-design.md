# ccstatuswidgets Phase 3 — Styling & Formatting Design Spec

## Overview

Phase 3 adds per-widget color customization (foreground, background), text formatting (bold, dim, italic, underline), configurable widget separators, and Powerline mode with Nerd Font arrow separators.

## Goals

1. Per-widget foreground and background colors (hex, 256-color, named)
2. Per-widget text formatting (bold, dim, italic, underline)
3. Configurable separator between widgets
4. Powerline mode with arrow glyphs
5. Interactive color picking in `ccw configure`
6. Backward compatible — existing configs work unchanged

## Non-Goals

- Font type/size (controlled by terminal emulator, not the binary)
- Built-in themes (users configure colors directly)
- Windows support

---

## Color System

### Supported Color Formats

| Format | Example | ANSI Output |
|--------|---------|-------------|
| Named | `"red"` | `\033[0;31m` (existing behavior) |
| 256-color | `"196"` or `"color:196"` | `\033[38;5;196m` |
| Hex truecolor | `"#ff6b6b"` | `\033[38;2;255;107;107m` |

Background colors use the same formats with `bg` prefix in ANSI:
- Named bg: `\033[41m` (red background)
- 256-color bg: `\033[48;5;196m`
- Hex bg: `\033[48;2;255;107;107m`

### Color Resolution Priority

When rendering a widget:
1. Check per-widget `fg`/`bg` in config → use if set
2. Fall back to widget's default `Color` field from `WidgetOutput`
3. If still empty, render plain (no color)

---

## Config Schema Changes

### Current (Phase 2)

```json
{
  "timeout_ms": 500,
  "lines": [...],
  "widgets": {
    "model": {},
    "weather": {"city": "Colombo"}
  }
}
```

### Phase 3 Additions

```json
{
  "timeout_ms": 500,
  "separator": " │ ",
  "powerline": false,
  "lines": [...],
  "widgets": {
    "model": {
      "fg": "#ff79c6",
      "bg": "#282a36",
      "bold": true
    },
    "directory": {
      "fg": "#8be9fd",
      "italic": true
    },
    "weather": {
      "city": "Colombo",
      "fg": "#f1fa8c"
    }
  }
}
```

### New Top-Level Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `separator` | string | `" "` | Character(s) between widgets on the same line |
| `powerline` | bool | `false` | Enable Powerline arrow separators |

### New Per-Widget Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fg` | string | (widget default) | Foreground color (named, `"196"`, or `"#hex"`) |
| `bg` | string | (none) | Background color |
| `bold` | bool | `false` | Bold text |
| `dim` | bool | `false` | Dim/faint text |
| `italic` | bool | `false` | Italic text |
| `underline` | bool | `false` | Underline text |

These mix with existing widget-specific config (city, symbols, etc.) in the same map.

---

## Renderer Changes

### Current Renderer

The renderer currently does:
1. For each widget result, if `Color` is set, wrap text in named ANSI color + reset
2. Join widgets with space
3. Join lines with `\n`

### New Renderer

1. For each widget result:
   a. Determine foreground: `cfg.fg` → `WidgetOutput.Color` → none
   b. Determine background: `cfg.bg` → none
   c. Determine formatting: bold/dim/italic/underline from cfg
   d. Build ANSI prefix from all active styles
   e. Wrap text: `prefix + text + reset`
2. Join widgets with configured separator
3. If powerline mode: use arrow separators with proper bg color transitions
4. Join lines with `\n`

### ANSI Code Building

```go
func buildANSI(fg, bg string, bold, dim, italic, underline bool) string {
    var codes []string
    if bold { codes = append(codes, "1") }
    if dim { codes = append(codes, "2") }
    if italic { codes = append(codes, "3") }
    if underline { codes = append(codes, "4") }
    if fg != "" { codes = append(codes, fgCode(fg)) }
    if bg != "" { codes = append(codes, bgCode(bg)) }
    if len(codes) == 0 { return "" }
    return "\033[" + strings.Join(codes, ";") + "m"
}
```

### Color Parsing

```go
func fgCode(color string) string {
    // Named: "red" → "31"
    // 256: "196" or "color:196" → "38;5;196"
    // Hex: "#ff6b6b" → "38;2;255;107;107"
}

func bgCode(color string) string {
    // Named: "red" → "41"
    // 256: "196" → "48;5;196"
    // Hex: "#ff6b6b" → "48;2;255;107;107"
}
```

---

## Powerline Mode

When `"powerline": true`:

- Widgets are rendered with background colors
- Between widgets, insert arrow separator: `` (U+E0B0) or `` (U+E0B2)
- Arrow fg = previous widget's bg, arrow bg = next widget's bg
- First widget gets a left cap, last widget gets a right cap
- Requires a Nerd Font installed in the terminal

### Example Output

```
 model  effort  directory  git-branch
```

### Fallback

If powerline is enabled but widgets don't have `bg` set, auto-assign alternating background colors from a default palette.

---

## Separator Config

### Simple separator (non-powerline)

```json
"separator": " │ "
```

Renders: `model │ effort │ directory │ git-branch`

### Common separator values

| Value | Result |
|-------|--------|
| `" "` | `model effort directory` (default, current behavior) |
| `" │ "` | `model │ effort │ directory` |
| `" · "` | `model · effort · directory` |
| `" ▸ "` | `model ▸ effort ▸ directory` |
| `" ┃ "` | `model ┃ effort ┃ directory` |

---

## Interactive Color Picker (ccw configure)

Add color editing to the existing configurator. When cursor is on a widget:

- Press `c` to edit foreground color
- Press `b` to edit background color
- Press `f` to toggle formatting (bold/dim/italic/underline)

### Color picker flow:

```
  Set foreground color for: model

  Named:   1. red  2. green  3. yellow  4. blue  5. magenta  6. cyan  7. white  8. dim  9. gray
  Custom:  Type hex (#ff6b6b) or 256-color number (196)
  Clear:   0 to remove custom color

  Enter choice:
```

### Formatting toggle flow:

```
  Formatting for: model

  [x] bold      (b)
  [ ] dim       (d)
  [x] italic    (i)
  [ ] underline (u)

  Press key to toggle, Enter to confirm
```

---

## Files Changed

### New Files
- `internal/renderer/colors.go` — color parsing (named, 256, hex → ANSI codes)
- `internal/renderer/colors_test.go`
- `internal/renderer/style.go` — style builder (fg + bg + bold/dim/italic/underline → ANSI prefix)
- `internal/renderer/style_test.go`

### Modified Files
- `internal/renderer/renderer.go` — use new style system, separator, powerline
- `internal/renderer/renderer_test.go` — new tests for styling
- `internal/config/config.go` — add `Separator` and `Powerline` fields to Config
- `internal/config/defaults.go` — default separator = `" "`
- `internal/cli/configure.go` — add color/formatting editing (c, b, f keys)

---

## Backward Compatibility

- Existing configs with no `fg`/`bg`/`bold` fields work exactly as before
- The default `Color` field from `WidgetOutput` is still respected
- Default separator is `" "` (space) — current behavior
- Powerline defaults to `false`
- Raw ANSI in widget text (lines-changed, stocks) is still passed through

---

## Testing

- Color parsing: named → ANSI, 256 → ANSI, hex → ANSI, invalid → empty
- Style building: combinations of fg + bg + bold + italic
- Renderer with custom separator
- Renderer with per-widget fg/bg overrides
- Powerline arrow separator generation
- Backward compat: existing test cases still pass
