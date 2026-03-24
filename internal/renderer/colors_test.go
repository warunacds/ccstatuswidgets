package renderer

import "testing"

func TestFgCode_NamedColors(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"red", "31"},
		{"green", "32"},
		{"yellow", "33"},
		{"blue", "34"},
		{"magenta", "35"},
		{"cyan", "36"},
		{"white", "37"},
		{"dim", "2"},
		{"gray", "90"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FgCode(tc.name)
			if got != tc.want {
				t.Errorf("FgCode(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestBgCode_NamedColors(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"red", "41"},
		{"green", "42"},
		{"yellow", "43"},
		{"blue", "44"},
		{"magenta", "45"},
		{"cyan", "46"},
		{"white", "47"},
		{"dim", ""},
		{"gray", "100"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BgCode(tc.name)
			if got != tc.want {
				t.Errorf("BgCode(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestFgCode_256Color_Numeric(t *testing.T) {
	got := FgCode("196")
	want := "38;5;196"
	if got != want {
		t.Errorf("FgCode(%q) = %q, want %q", "196", got, want)
	}
}

func TestBgCode_256Color_Numeric(t *testing.T) {
	got := BgCode("196")
	want := "48;5;196"
	if got != want {
		t.Errorf("BgCode(%q) = %q, want %q", "196", got, want)
	}
}

func TestFgCode_256Color_Prefixed(t *testing.T) {
	got := FgCode("color:42")
	want := "38;5;42"
	if got != want {
		t.Errorf("FgCode(%q) = %q, want %q", "color:42", got, want)
	}
}

func TestFgCode_Hex(t *testing.T) {
	got := FgCode("#ff6b6b")
	want := "38;2;255;107;107"
	if got != want {
		t.Errorf("FgCode(%q) = %q, want %q", "#ff6b6b", got, want)
	}
}

func TestBgCode_Hex(t *testing.T) {
	got := BgCode("#ff6b6b")
	want := "48;2;255;107;107"
	if got != want {
		t.Errorf("BgCode(%q) = %q, want %q", "#ff6b6b", got, want)
	}
}

func TestFgCode_Hex_Black(t *testing.T) {
	got := FgCode("#000000")
	want := "38;2;0;0;0"
	if got != want {
		t.Errorf("FgCode(%q) = %q, want %q", "#000000", got, want)
	}
}

func TestBgCode_Hex_Black(t *testing.T) {
	got := BgCode("#000000")
	want := "48;2;0;0;0"
	if got != want {
		t.Errorf("BgCode(%q) = %q, want %q", "#000000", got, want)
	}
}

func TestFgCode_Empty(t *testing.T) {
	got := FgCode("")
	if got != "" {
		t.Errorf("FgCode(%q) = %q, want %q", "", got, "")
	}
}

func TestBgCode_Empty(t *testing.T) {
	got := BgCode("")
	if got != "" {
		t.Errorf("BgCode(%q) = %q, want %q", "", got, "")
	}
}

func TestFgCode_Invalid(t *testing.T) {
	got := FgCode("notacolor")
	if got != "" {
		t.Errorf("FgCode(%q) = %q, want %q", "notacolor", got, "")
	}
}

func TestBgCode_Invalid(t *testing.T) {
	got := BgCode("notacolor")
	if got != "" {
		t.Errorf("BgCode(%q) = %q, want %q", "notacolor", got, "")
	}
}

func TestFgCode_Hex_Uppercase(t *testing.T) {
	got := FgCode("#FF6B6B")
	want := "38;2;255;107;107"
	if got != want {
		t.Errorf("FgCode(%q) = %q, want %q", "#FF6B6B", got, want)
	}
}

func TestBgCode_Hex_Uppercase(t *testing.T) {
	got := BgCode("#FF6B6B")
	want := "48;2;255;107;107"
	if got != want {
		t.Errorf("BgCode(%q) = %q, want %q", "#FF6B6B", got, want)
	}
}
