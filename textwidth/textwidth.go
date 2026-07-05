// Package textwidth measures the terminal display width of runes and
// strings. Wide characters (CJK, emoji) occupy two cells; combining
// marks and zero-width characters occupy none.
//
// The tables cover the common East Asian Wide/Fullwidth blocks and the
// major emoji ranges. Ambiguous-width characters are treated as narrow,
// matching most modern terminal defaults.
package textwidth

import (
	"strings"
	"unicode"
)

// interval is an inclusive rune range.
type interval struct {
	first, last rune
}

// wide holds rune ranges that render as two cells, sorted ascending.
var wide = []interval{
	{0x1100, 0x115F},   // Hangul Jamo
	{0x231A, 0x231B},   // watch, hourglass
	{0x2329, 0x232A},   // angle brackets
	{0x23E9, 0x23EC},   // media control symbols
	{0x23F0, 0x23F0},   // alarm clock
	{0x23F3, 0x23F3},   // hourglass with flowing sand
	{0x25FD, 0x25FE},   // small squares
	{0x2614, 0x2615},   // umbrella, hot beverage
	{0x2648, 0x2653},   // zodiac
	{0x267F, 0x267F},   // wheelchair
	{0x2693, 0x2693},   // anchor
	{0x26A1, 0x26A1},   // high voltage
	{0x26AA, 0x26AB},   // circles
	{0x26BD, 0x26BE},   // soccer, baseball
	{0x26C4, 0x26C5},   // snowman, sun behind cloud
	{0x26CE, 0x26CE},   // ophiuchus
	{0x26D4, 0x26D4},   // no entry
	{0x26EA, 0x26EA},   // church
	{0x26F2, 0x26F3},   // fountain, flag in hole
	{0x26F5, 0x26F5},   // sailboat
	{0x26FA, 0x26FA},   // tent
	{0x26FD, 0x26FD},   // fuel pump
	{0x2705, 0x2705},   // check mark
	{0x270A, 0x270B},   // fists
	{0x2728, 0x2728},   // sparkles
	{0x274C, 0x274C},   // cross mark
	{0x274E, 0x274E},   // cross mark button
	{0x2753, 0x2755},   // question/exclamation marks
	{0x2757, 0x2757},   // exclamation
	{0x2795, 0x2797},   // plus, minus, divide
	{0x27B0, 0x27B0},   // curly loop
	{0x27BF, 0x27BF},   // double curly loop
	{0x2B1B, 0x2B1C},   // large squares
	{0x2B50, 0x2B50},   // star
	{0x2B55, 0x2B55},   // heavy circle
	{0x2E80, 0x303E},   // CJK radicals, Kangxi, CJK symbols & punctuation
	{0x3041, 0x33FF},   // Hiragana .. CJK compatibility
	{0x3400, 0x4DBF},   // CJK ext A
	{0x4E00, 0x9FFF},   // CJK unified
	{0xA000, 0xA4CF},   // Yi
	{0xA960, 0xA97F},   // Hangul Jamo ext A
	{0xAC00, 0xD7A3},   // Hangul syllables
	{0xF900, 0xFAFF},   // CJK compatibility ideographs
	{0xFE10, 0xFE19},   // vertical forms
	{0xFE30, 0xFE6F},   // CJK compatibility forms, small form variants
	{0xFF00, 0xFF60},   // fullwidth forms
	{0xFFE0, 0xFFE6},   // fullwidth signs
	{0x16FE0, 0x16FE4}, // Tangut/ideographic marks
	{0x17000, 0x187F7}, // Tangut
	{0x18800, 0x18CD5}, // Tangut components
	{0x1AFF0, 0x1B16F}, // Katakana extensions
	{0x1F004, 0x1F004}, // mahjong red dragon
	{0x1F0CF, 0x1F0CF}, // joker
	{0x1F18E, 0x1F18E}, // AB button
	{0x1F191, 0x1F19A}, // squared symbols
	{0x1F200, 0x1F2FF}, // enclosed ideographic supplement
	{0x1F300, 0x1F64F}, // misc symbols & pictographs, emoticons
	{0x1F680, 0x1F6FF}, // transport & map symbols
	{0x1F7E0, 0x1F7EB}, // large colored circles/squares
	{0x1F90C, 0x1F9FF}, // supplemental symbols & pictographs
	{0x1FA70, 0x1FAFF}, // symbols & pictographs ext A
	{0x20000, 0x2FFFD}, // CJK ext B..
	{0x30000, 0x3FFFD}, // CJK ext G..
}

// zeroWidth holds rune ranges that occupy no cells beyond combining
// marks (covered by unicode.Mn/Me), sorted ascending.
var zeroWidth = []interval{
	{0x200B, 0x200F}, // zero-width space/joiners, directional marks
	{0x2028, 0x202E}, // line/paragraph separators, directional overrides
	{0x2060, 0x2064}, // word joiner, invisible operators
	{0xFE00, 0xFE0F}, // variation selectors
	{0xFEFF, 0xFEFF}, // BOM / zero-width no-break space
}

func in(r rune, table []interval) bool {
	lo, hi := 0, len(table)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		iv := table[mid]
		switch {
		case r < iv.first:
			hi = mid - 1
		case r > iv.last:
			lo = mid + 1
		default:
			return true
		}
	}
	return false
}

// Rune returns the display width of a rune: 0, 1, or 2 cells.
func Rune(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 0x20 || (r >= 0x7F && r < 0xA0):
		return 0 // control characters
	case r < 0x300:
		return 1 // fast path for ASCII and Latin-1
	case unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r):
		return 0 // combining marks
	case in(r, zeroWidth):
		return 0
	case in(r, wide):
		return 2
	default:
		return 1
	}
}

// String returns the total display width of a string in cells.
func String(s string) int {
	w := 0
	for _, r := range s {
		w += Rune(r)
	}
	return w
}

// SplitLines splits s on newlines, expanding each tab to a single
// space. Tabs have no well-defined cell width in the renderer, so they
// are normalized before measurement to keep layout and paint agreeing.
func SplitLines(s string) []string {
	s = strings.ReplaceAll(s, "\t", " ")
	return strings.Split(s, "\n")
}

// Wrap word-wraps s to fit maxWidth display cells, preserving leading
// whitespace on wrapped continuation lines. Lines that already fit are
// kept verbatim (internal spacing intact). A single word wider than
// maxWidth is left unbroken — the renderer clips it.
func Wrap(s string, maxWidth int) []string {
	if maxWidth <= 0 {
		return nil
	}
	if s == "" {
		return []string{""}
	}
	var lines []string
	for _, raw := range SplitLines(s) {
		if String(raw) <= maxWidth {
			lines = append(lines, raw)
			continue
		}
		trimmed := strings.TrimLeft(raw, " ")
		leading := raw[:len(raw)-len(trimmed)]
		words := strings.Fields(trimmed)
		if len(words) == 0 {
			lines = append(lines, raw)
			continue
		}
		line := leading + words[0]
		lineLen := String(line)
		for _, w := range words[1:] {
			wLen := String(w)
			if lineLen+1+wLen <= maxWidth {
				line += " " + w
				lineLen += 1 + wLen
			} else {
				lines = append(lines, line)
				line = leading + w
				lineLen = String(leading) + wLen
			}
		}
		lines = append(lines, line)
	}
	return lines
}

// Truncate cuts s so its display width does not exceed maxWidth,
// appending the ellipsis rune when truncation occurs.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if String(s) <= maxWidth {
		return s
	}
	// Reserve one cell for the ellipsis.
	w := 0
	out := make([]rune, 0, len(s))
	for _, r := range s {
		rw := Rune(r)
		if w+rw > maxWidth-1 {
			break
		}
		out = append(out, r)
		w += rw
	}
	return string(out) + "…"
}
