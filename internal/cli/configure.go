package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

// RunConfigure runs the interactive widget configurator TUI.
func RunConfigure(configDir string, allWidgetNames []string) error {
	sort.Strings(allWidgetNames)

	configPath := filepath.Join(configDir, "config.json")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Deep-copy the lines so we can detect changes.
	originalLines := deepCopyLines(cfg.Lines)

	// Open /dev/tty directly for input (not os.Stdin, since ccw might be in pipeline mode).
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer tty.Close()

	// Set terminal to raw mode.
	if err := setRawMode(); err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer restoreTerminal()

	// Handle Ctrl+C: restore terminal before exiting.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		restoreTerminal()
		os.Exit(0)
	}()

	state := &configureState{
		cfg:            cfg,
		allWidgetNames: allWidgetNames,
		cursorRow:      0,
		cursorCol:      0,
		moveMode:       false,
		dirty:          false,
		tty:            tty,
		configPath:     configPath,
		originalLines:  originalLines,
	}

	// Ensure at least one row exists.
	if len(state.cfg.Lines) == 0 {
		state.cfg.Lines = append(state.cfg.Lines, config.LineConfig{Widgets: []string{}})
		state.dirty = true
	}

	state.render()

	reader := bufio.NewReader(tty)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}

		switch b {
		case 3: // Ctrl+C
			restoreTerminal()
			return nil

		case 27: // ESC sequence (arrow keys)
			b2, err := reader.ReadByte()
			if err != nil {
				continue
			}
			if b2 != '[' {
				continue
			}
			b3, err := reader.ReadByte()
			if err != nil {
				continue
			}
			switch b3 {
			case 'A': // Up
				state.moveUp()
			case 'B': // Down
				state.moveDown()
			case 'C': // Right
				state.moveRight()
			case 'D': // Left
				state.moveLeft()
			}

		case 'a': // Add widget
			state.addWidget(reader)

		case 'd': // Delete widget
			state.deleteWidget()

		case 'm': // Toggle move mode
			state.toggleMoveMode()

		case 13: // Enter — confirm move mode
			if state.moveMode {
				state.moveMode = false
			}

		case 'n': // New row
			state.newRow()

		case 'r': // Remove row (if empty)
			state.removeRow()

		case 's': // Save and quit
			restoreTerminal()
			if err := config.Save(configPath, cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Print("\033[2J\033[H")
			fmt.Println("Configuration saved.")
			return nil

		case 'q': // Quit without saving
			if state.dirty {
				state.renderMessage("Unsaved changes. Press q again to quit, or s to save.")
				b2, err := reader.ReadByte()
				if err != nil {
					break
				}
				if b2 == 'q' {
					restoreTerminal()
					fmt.Print("\033[2J\033[H")
					fmt.Println("Quit without saving.")
					return nil
				} else if b2 == 's' {
					restoreTerminal()
					if err := config.Save(configPath, cfg); err != nil {
						return fmt.Errorf("failed to save config: %w", err)
					}
					fmt.Print("\033[2J\033[H")
					fmt.Println("Configuration saved.")
					return nil
				}
				// Any other key: cancel quit, re-render.
			} else {
				restoreTerminal()
				fmt.Print("\033[2J\033[H")
				return nil
			}
		}

		state.render()
	}

	return nil
}

type configureState struct {
	cfg            *config.Config
	allWidgetNames []string
	cursorRow      int
	cursorCol      int
	moveMode       bool
	dirty          bool
	tty            *os.File
	configPath     string
	originalLines  []config.LineConfig
	message        string
}

func (s *configureState) availableWidgets() []string {
	// Build set of widgets already in config.
	used := make(map[string]bool)
	for _, line := range s.cfg.Lines {
		for _, w := range line.Widgets {
			used[w] = true
		}
	}

	var available []string
	for _, name := range s.allWidgetNames {
		if !used[name] {
			available = append(available, name)
		}
	}
	return available
}

func (s *configureState) clampCursor() {
	if len(s.cfg.Lines) == 0 {
		s.cursorRow = 0
		s.cursorCol = 0
		return
	}
	if s.cursorRow >= len(s.cfg.Lines) {
		s.cursorRow = len(s.cfg.Lines) - 1
	}
	if s.cursorRow < 0 {
		s.cursorRow = 0
	}
	rowLen := len(s.cfg.Lines[s.cursorRow].Widgets)
	if rowLen == 0 {
		s.cursorCol = 0
	} else if s.cursorCol >= rowLen {
		s.cursorCol = rowLen - 1
	}
	if s.cursorCol < 0 {
		s.cursorCol = 0
	}
}

func (s *configureState) moveUp() {
	if s.cursorRow > 0 {
		s.cursorRow--
		s.clampCursor()
	}
}

func (s *configureState) moveDown() {
	if s.cursorRow < len(s.cfg.Lines)-1 {
		s.cursorRow++
		s.clampCursor()
	}
}

func (s *configureState) moveRight() {
	if len(s.cfg.Lines) == 0 {
		return
	}
	row := s.cfg.Lines[s.cursorRow].Widgets
	if len(row) == 0 {
		return
	}

	if s.moveMode {
		// Move the widget right.
		if s.cursorCol < len(row)-1 {
			row[s.cursorCol], row[s.cursorCol+1] = row[s.cursorCol+1], row[s.cursorCol]
			s.cursorCol++
			s.dirty = true
		}
	} else {
		if s.cursorCol < len(row)-1 {
			s.cursorCol++
		}
	}
}

func (s *configureState) moveLeft() {
	if len(s.cfg.Lines) == 0 {
		return
	}
	row := s.cfg.Lines[s.cursorRow].Widgets
	if len(row) == 0 {
		return
	}

	if s.moveMode {
		// Move the widget left.
		if s.cursorCol > 0 {
			row[s.cursorCol], row[s.cursorCol-1] = row[s.cursorCol-1], row[s.cursorCol]
			s.cursorCol--
			s.dirty = true
		}
	} else {
		if s.cursorCol > 0 {
			s.cursorCol--
		}
	}
}

func (s *configureState) deleteWidget() {
	if len(s.cfg.Lines) == 0 {
		return
	}
	row := s.cfg.Lines[s.cursorRow].Widgets
	if len(row) == 0 {
		return
	}

	// Remove widget at cursor. Build a new slice to avoid mutating the original backing array.
	newRow := make([]string, 0, len(row)-1)
	newRow = append(newRow, row[:s.cursorCol]...)
	newRow = append(newRow, row[s.cursorCol+1:]...)
	s.cfg.Lines[s.cursorRow].Widgets = newRow
	s.dirty = true
	s.clampCursor()
}

func (s *configureState) toggleMoveMode() {
	if len(s.cfg.Lines) == 0 {
		return
	}
	if len(s.cfg.Lines[s.cursorRow].Widgets) == 0 {
		return
	}
	s.moveMode = !s.moveMode
}

func (s *configureState) newRow() {
	// Insert a new empty row below the current row.
	insertIdx := s.cursorRow + 1
	newLines := make([]config.LineConfig, 0, len(s.cfg.Lines)+1)
	newLines = append(newLines, s.cfg.Lines[:insertIdx]...)
	newLines = append(newLines, config.LineConfig{Widgets: []string{}})
	newLines = append(newLines, s.cfg.Lines[insertIdx:]...)
	s.cfg.Lines = newLines
	s.cursorRow = insertIdx
	s.cursorCol = 0
	s.dirty = true
}

func (s *configureState) removeRow() {
	if len(s.cfg.Lines) == 0 {
		return
	}
	if len(s.cfg.Lines[s.cursorRow].Widgets) > 0 {
		s.message = "Cannot remove non-empty row. Delete all widgets first."
		return
	}
	if len(s.cfg.Lines) <= 1 {
		s.message = "Cannot remove the last row."
		return
	}

	newLines := make([]config.LineConfig, 0, len(s.cfg.Lines)-1)
	newLines = append(newLines, s.cfg.Lines[:s.cursorRow]...)
	newLines = append(newLines, s.cfg.Lines[s.cursorRow+1:]...)
	s.cfg.Lines = newLines
	s.dirty = true
	s.clampCursor()
}

func (s *configureState) addWidget(reader *bufio.Reader) {
	available := s.availableWidgets()
	if len(available) == 0 {
		s.message = "No available widgets to add."
		return
	}

	// Show available widgets.
	s.renderAddMenu(available)

	// Read user input: number + Enter.
	// We need to read a line. Since we are in raw mode, read chars until Enter.
	var input []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return
		}
		if b == 13 || b == 10 { // Enter
			break
		}
		if b == 27 || b == 'q' { // ESC or q to cancel
			s.message = "Add cancelled."
			return
		}
		if b == 127 || b == 8 { // Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
				// Re-render the prompt with updated input.
				fmt.Printf("\r\033[K  Enter number (or q to cancel): %s", string(input))
			}
			continue
		}
		if b >= '0' && b <= '9' {
			input = append(input, b)
			fmt.Printf("%c", b)
		}
	}

	if len(input) == 0 {
		s.message = "Add cancelled."
		return
	}

	num, err := strconv.Atoi(string(input))
	if err != nil || num < 1 || num > len(available) {
		s.message = fmt.Sprintf("Invalid selection: %s", string(input))
		return
	}

	widgetName := available[num-1]

	// Insert at cursor position.
	row := s.cfg.Lines[s.cursorRow].Widgets
	insertIdx := s.cursorCol
	if len(row) == 0 {
		insertIdx = 0
	}

	newRow := make([]string, 0, len(row)+1)
	newRow = append(newRow, row[:insertIdx]...)
	newRow = append(newRow, widgetName)
	newRow = append(newRow, row[insertIdx:]...)
	s.cfg.Lines[s.cursorRow].Widgets = newRow
	s.cursorCol = insertIdx
	s.dirty = true
	s.message = fmt.Sprintf("Added %s", widgetName)
}

func (s *configureState) render() {
	var b strings.Builder

	// Clear screen, move to top-left.
	b.WriteString("\033[2J\033[H")

	dirtyMarker := ""
	if s.dirty {
		dirtyMarker = " *"
	}
	b.WriteString(fmt.Sprintf("  ccstatuswidgets configurator%s\n", dirtyMarker))
	b.WriteString("  ─────────────────────────────\n\n")

	// Render rows — each widget on its own line within the row.
	for i, line := range s.cfg.Lines {
		isCurrentRow := i == s.cursorRow
		rowMarker := "  "
		if isCurrentRow {
			rowMarker = "▸ "
		}
		b.WriteString(fmt.Sprintf("%sRow %d:\n", rowMarker, i+1))

		if len(line.Widgets) == 0 {
			if isCurrentRow {
				b.WriteString("    \033[2m(empty — press 'a' to add)\033[0m\n")
			} else {
				b.WriteString("    \033[2m(empty)\033[0m\n")
			}
		} else {
			for j, w := range line.Widgets {
				prefix := "    "
				if isCurrentRow && j == s.cursorCol {
					if s.moveMode {
						b.WriteString(fmt.Sprintf("%s\033[33m▶ <<%s>>\033[0m\n", prefix, w))
					} else {
						b.WriteString(fmt.Sprintf("%s\033[36m▶ %s\033[0m\n", prefix, w))
					}
				} else {
					b.WriteString(fmt.Sprintf("%s  %s\n", prefix, w))
				}
			}
		}
		b.WriteString("\n")
	}

	// Available widgets.
	available := s.availableWidgets()
	b.WriteString("  Available:\n")
	if len(available) > 0 {
		// Show in columns of 3
		for i := 0; i < len(available); i += 3 {
			b.WriteString("    ")
			for j := i; j < i+3 && j < len(available); j++ {
				b.WriteString(fmt.Sprintf("%-20s", available[j]))
			}
			b.WriteString("\n")
		}
	} else {
		b.WriteString("    \033[2m(all widgets enabled)\033[0m\n")
	}

	// Controls.
	b.WriteString("\n  ─────────────────────────────\n")
	if s.moveMode {
		b.WriteString("  ←→ move widget  m/Enter confirm  \n")
	} else {
		b.WriteString("  ↑↓ row  ←→ widget  a add  d delete  m move\n")
		b.WriteString("  n new row  r remove row  s save  q quit\n")
	}

	// Status line.
	b.WriteString("  ─────────────────────────────\n")
	if len(s.cfg.Lines) > 0 && len(s.cfg.Lines[s.cursorRow].Widgets) > 0 {
		widgetName := s.cfg.Lines[s.cursorRow].Widgets[s.cursorCol]
		modeLabel := ""
		if s.moveMode {
			modeLabel = " \033[33mMOVE\033[0m"
		}
		b.WriteString(fmt.Sprintf("  Row %d, Col %d: \033[1m%s\033[0m%s\n", s.cursorRow+1, s.cursorCol+1, widgetName, modeLabel))
	} else {
		b.WriteString(fmt.Sprintf("  Row %d (empty)\n", s.cursorRow+1))
	}

	// Message line.
	if s.message != "" {
		b.WriteString(fmt.Sprintf("\n  \033[33m%s\033[0m\n", s.message))
		s.message = ""
	}

	fmt.Print(b.String())
}

func (s *configureState) renderMessage(msg string) {
	s.message = msg
	s.render()
}

func (s *configureState) renderAddMenu(available []string) {
	var b strings.Builder
	b.WriteString("\033[2J\033[H")
	b.WriteString("Add widget\n\n")
	for i, name := range available {
		b.WriteString(fmt.Sprintf("  %2d. %s\n", i+1, name))
	}
	b.WriteString("\n  Enter number (or q to cancel): ")
	fmt.Print(b.String())
}

// deepCopyLines creates a deep copy of config lines for change detection.
func deepCopyLines(lines []config.LineConfig) []config.LineConfig {
	cp := make([]config.LineConfig, len(lines))
	for i, l := range lines {
		cp[i] = config.LineConfig{
			Widgets: make([]string, len(l.Widgets)),
		}
		copy(cp[i].Widgets, l.Widgets)
	}
	return cp
}

// setRawMode puts the terminal into raw mode using stty.
func setRawMode() error {
	flag := "-f"
	if runtime.GOOS == "linux" {
		flag = "-F"
	}
	return exec.Command("stty", flag, "/dev/tty", "raw", "-echo").Run()
}

// restoreTerminal restores the terminal to sane mode.
func restoreTerminal() {
	flag := "-f"
	if runtime.GOOS == "linux" {
		flag = "-F"
	}
	exec.Command("stty", flag, "/dev/tty", "sane").Run()
}
