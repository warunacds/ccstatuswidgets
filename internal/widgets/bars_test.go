package widgets

import "testing"

func TestBuildBar_30Percent(t *testing.T) {
	bar := BuildBar(30, 10)
	filled := "\u2588\u2588\u2588"
	empty := "\u2591\u2591\u2591\u2591\u2591\u2591\u2591"
	expected := filled + empty
	if bar != expected {
		t.Errorf("BuildBar(30, 10) = %q, want %q", bar, expected)
	}
}

func TestBuildBar_ZeroPercent(t *testing.T) {
	bar := BuildBar(0, 10)
	expected := "\u2591\u2591\u2591\u2591\u2591\u2591\u2591\u2591\u2591\u2591"
	if bar != expected {
		t.Errorf("BuildBar(0, 10) = %q, want %q", bar, expected)
	}
}

func TestBuildBar_100Percent(t *testing.T) {
	bar := BuildBar(100, 10)
	expected := "\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588"
	if bar != expected {
		t.Errorf("BuildBar(100, 10) = %q, want %q", bar, expected)
	}
}

func TestBarColor_Green(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{0, "green"},
		{49, "green"},
	}
	for _, tt := range tests {
		got := BarColor(tt.pct)
		if got != tt.want {
			t.Errorf("BarColor(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

func TestBarColor_Yellow(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{50, "yellow"},
		{79, "yellow"},
	}
	for _, tt := range tests {
		got := BarColor(tt.pct)
		if got != tt.want {
			t.Errorf("BarColor(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

func TestBarColor_Red(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{80, "red"},
		{100, "red"},
	}
	for _, tt := range tests {
		got := BarColor(tt.pct)
		if got != tt.want {
			t.Errorf("BarColor(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}
