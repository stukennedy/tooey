package ansi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stukennedy/tooey/cell"
	"github.com/stukennedy/tooey/diff"
	"github.com/stukennedy/tooey/node"
)

func renderCells(cells ...cell.Cell) string {
	var buf bytes.Buffer
	Render(&buf, []diff.Change{{X: 0, Y: 0, Cells: cells}})
	return buf.String()
}

func TestRenderRGBTrueColor(t *testing.T) {
	SetTrueColor(true)
	defer SetTrueColor(detectTrueColor())

	out := renderCells(cell.Cell{Rune: 'A', FG: node.RGB(255, 128, 0), BG: node.RGB(0, 0, 64)})
	if !strings.Contains(out, ";38;2;255;128;0") {
		t.Fatalf("missing RGB foreground SGR, got %q", out)
	}
	if !strings.Contains(out, ";48;2;0;0;64") {
		t.Fatalf("missing RGB background SGR, got %q", out)
	}
}

func TestRenderRGBDowngrade(t *testing.T) {
	SetTrueColor(false)
	defer SetTrueColor(detectTrueColor())

	// Pure red should downgrade to cube entry 196
	out := renderCells(cell.Cell{Rune: 'A', FG: node.RGB(255, 0, 0)})
	if !strings.Contains(out, ";38;5;196m") {
		t.Fatalf("expected downgrade to palette 196, got %q", out)
	}
	if strings.Contains(out, ";38;2;") {
		t.Fatalf("should not emit 24-bit SGR when truecolor disabled, got %q", out)
	}
}

func TestRenderExplicitAnsiBlack(t *testing.T) {
	out := renderCells(cell.Cell{Rune: 'A', FG: node.Ansi(0)})
	if !strings.Contains(out, ";38;5;0m") {
		t.Fatalf("Ansi(0) should emit palette black, got %q", out)
	}
}

func TestColorHelpers(t *testing.T) {
	if !node.Color(0).IsDefault() {
		t.Fatal("zero Color should be default")
	}
	if node.Color(245).IsDefault() || node.Color(245).IsRGB() {
		t.Fatal("plain palette color misclassified")
	}
	if node.Color(245).Ansi256() != 245 {
		t.Fatal("plain palette color should round-trip")
	}
	c := node.RGB(10, 20, 30)
	if !c.IsRGB() || c.IsDefault() {
		t.Fatal("RGB color misclassified")
	}
	r, g, b := c.RGBValues()
	if r != 10 || g != 20 || b != 30 {
		t.Fatalf("RGBValues round-trip failed: %d %d %d", r, g, b)
	}
	if node.Ansi(0).IsDefault() {
		t.Fatal("Ansi(0) must be distinct from default")
	}
	if node.Ansi(0).Ansi256() != 0 {
		t.Fatal("Ansi(0) should map to palette index 0")
	}
}

func TestRGBToAnsi256Grayscale(t *testing.T) {
	// Mid-gray lands on the grayscale ramp (232..255)
	got := node.RGB(128, 128, 128).Ansi256()
	if got < 232 {
		t.Fatalf("mid-gray should map to grayscale ramp, got %d", got)
	}
	if node.RGB(0, 0, 0).Ansi256() != 16 {
		t.Fatalf("black should map to cube 16, got %d", node.RGB(0, 0, 0).Ansi256())
	}
	if node.RGB(255, 255, 255).Ansi256() != 231 {
		t.Fatalf("white should map to cube 231, got %d", node.RGB(255, 255, 255).Ansi256())
	}
}
