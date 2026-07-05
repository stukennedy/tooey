// Package tooeytest provides golden-frame testing helpers: render a
// node tree at a fixed size and assert on the resulting text frame.
//
//	got := tooeytest.RenderText(view(model), 40, 10)
//	tooeytest.AssertFrame(t, view(model), 40, 10, `
//	┌────┐
//	│ hi │
//	└────┘`)
package tooeytest

import (
	"strings"
	"testing"

	"github.com/stukennedy/tooey/cell"
	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

// Render lays out and paints a node tree into a cell buffer of the
// given size.
func Render(n node.Node, w, h int) *cell.Buffer {
	lt := layout.Layout(n, w, h)
	buf := cell.NewBuffer(w, h)
	cell.Paint(buf, lt)
	return buf
}

// RenderText renders a node tree and returns the frame as text: one
// line per row, trailing whitespace trimmed, trailing blank lines
// dropped. Wide-rune continuation cells are skipped, so CJK and emoji
// appear once.
func RenderText(n node.Node, w, h int) string {
	return BufferText(Render(n, w, h))
}

// BufferText converts a painted buffer to text (see RenderText).
func BufferText(buf *cell.Buffer) string {
	lines := make([]string, 0, buf.Height)
	for y := 0; y < buf.Height; y++ {
		var b strings.Builder
		for x := 0; x < buf.Width; x++ {
			c := buf.Get(x, y)
			if c.IsContinuation() {
				continue
			}
			b.WriteRune(c.Rune)
		}
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// AssertFrame renders a node tree and fails the test if the frame does
// not match want. Leading/trailing blank lines in want are ignored, and
// common leading tab indentation is stripped, so the expectation can be
// written as an indented raw string literal.
func AssertFrame(t testing.TB, n node.Node, w, h int, want string) {
	t.Helper()
	got := RenderText(n, w, h)
	if got != normalize(want) {
		t.Errorf("frame mismatch\n--- want ---\n%s\n--- got ---\n%s", normalize(want), got)
	}
}

// normalize strips leading/trailing blank lines and common leading tab
// indentation from a raw-string expectation.
func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	// Find the common leading tab prefix.
	prefix := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := 0
		for n < len(l) && l[n] == '\t' {
			n++
		}
		if prefix == -1 || n < prefix {
			prefix = n
		}
	}
	if prefix > 0 {
		for i, l := range lines {
			if len(l) >= prefix {
				lines[i] = l[prefix:]
			}
		}
	}
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}
