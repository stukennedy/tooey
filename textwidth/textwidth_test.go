package textwidth

import "testing"

func TestRune(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{'a', 1},
		{'Z', 1},
		{' ', 1},
		{'é', 1},
		{'…', 1},
		{'─', 1},
		{'日', 2},
		{'本', 2},
		{'語', 2},
		{'한', 2},
		{'カ', 2},
		{'，', 2},      // fullwidth comma
		{'🎉', 2},      // emoji
		{'🚀', 2},      // transport emoji
		{'😀', 2},      // emoticon
		{0x0301, 0},   // combining acute accent
		{0x200B, 0},   // zero-width space
		{0xFE0F, 0},   // variation selector
		{0x07, 0},     // control
		{0, 0},        // NUL / continuation marker
	}
	for _, c := range cases {
		if got := Rune(c.r); got != c.want {
			t.Errorf("Rune(%U) = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"hello", 5},
		{"日本語", 6},
		{"a日b", 4},
		{"é", 1}, // e + combining accent
		{"🎉🎉", 4},
	}
	for _, c := range cases {
		if got := String(c.s); got != c.want {
			t.Errorf("String(%q) = %d, want %d", c.s, got, c.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		s        string
		maxWidth int
		want     string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"hello", 0, ""},
		{"日本語", 6, "日本語"},
		{"日本語", 5, "日本…"},
		{"日本語", 4, "日…"}, // can't split 本, ellipsis after 日 (width 3)
		{"日本語", 2, "…"},
	}
	for _, c := range cases {
		if got := Truncate(c.s, c.maxWidth); got != c.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", c.s, c.maxWidth, got, c.want)
		}
	}
}
