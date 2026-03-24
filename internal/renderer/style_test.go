package renderer

import (
	"strings"
	"testing"
)

func TestStyleFgOnly(t *testing.T) {
	s := WidgetStyle{Fg: "red"}
	got := s.Prefix()
	want := "\033[31m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleBgOnly(t *testing.T) {
	s := WidgetStyle{Bg: "green"}
	got := s.Prefix()
	want := "\033[42m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleFgAndBg(t *testing.T) {
	s := WidgetStyle{Fg: "red", Bg: "green"}
	got := s.Prefix()
	want := "\033[31;42m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleBoldOnly(t *testing.T) {
	s := WidgetStyle{Bold: true}
	got := s.Prefix()
	want := "\033[1m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleItalicOnly(t *testing.T) {
	s := WidgetStyle{Italic: true}
	got := s.Prefix()
	want := "\033[3m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleFgBoldItalic(t *testing.T) {
	s := WidgetStyle{Fg: "red", Bold: true, Italic: true}
	got := s.Prefix()
	want := "\033[1;3;31m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleAllFormatting(t *testing.T) {
	s := WidgetStyle{
		Fg:        "red",
		Bg:        "blue",
		Bold:      true,
		Dim:       true,
		Italic:    true,
		Underline: true,
	}
	got := s.Prefix()

	// Must contain all codes.
	for _, code := range []string{"1", "2", "3", "4", "31", "44"} {
		if !strings.Contains(got, code) {
			t.Errorf("prefix %q missing code %q", got, code)
		}
	}

	// Check exact value: bold;dim;italic;underline;fg;bg
	want := "\033[1;2;3;4;31;44m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleNoStyles(t *testing.T) {
	s := WidgetStyle{}
	got := s.Prefix()
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
	if s.HasStyle() {
		t.Error("HasStyle() should be false for zero WidgetStyle")
	}
}

func TestStyleHexFgNamedBg(t *testing.T) {
	s := WidgetStyle{Fg: "#ff6b6b", Bg: "blue"}
	got := s.Prefix()
	// #ff6b6b → RGB(255,107,107) → "38;2;255;107;107"
	// blue bg → "44"
	want := "\033[38;2;255;107;107;44m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyle256Color(t *testing.T) {
	s := WidgetStyle{Fg: "196"}
	got := s.Prefix()
	want := "\033[38;5;196m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStyleHasStyleTrue(t *testing.T) {
	tests := []struct {
		name  string
		style WidgetStyle
	}{
		{"fg", WidgetStyle{Fg: "red"}},
		{"bg", WidgetStyle{Bg: "green"}},
		{"bold", WidgetStyle{Bold: true}},
		{"dim", WidgetStyle{Dim: true}},
		{"italic", WidgetStyle{Italic: true}},
		{"underline", WidgetStyle{Underline: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.style.HasStyle() {
				t.Errorf("HasStyle() should be true for %s", tc.name)
			}
		})
	}
}
