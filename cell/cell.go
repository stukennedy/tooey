package cell

import (
	"github.com/stukennedy/tooey/node"
	"github.com/stukennedy/tooey/textwidth"
)

// Cell represents a single terminal cell. A Rune of 0 marks the
// continuation half of a wide (two-cell) character in the previous cell.
type Cell struct {
	Rune  rune
	FG    node.Color
	BG    node.Color
	Style node.StyleFlags
}

// IsContinuation reports whether this cell is the right half of a wide
// character. Consumers reading the buffer should skip these cells — the
// wide rune to the left already spans both.
func (c Cell) IsContinuation() bool { return c.Rune == 0 }

// Buffer is a row-major flat cell buffer representing a terminal frame.
type Buffer struct {
	Width  int
	Height int
	Cells  []Cell
}

// NewBuffer creates a buffer filled with spaces.
func NewBuffer(w, h int) *Buffer {
	cells := make([]Cell, w*h)
	for i := range cells {
		cells[i].Rune = ' '
	}
	return &Buffer{Width: w, Height: h, Cells: cells}
}

// inBounds checks if coordinates are valid.
func (b *Buffer) inBounds(x, y int) bool {
	return x >= 0 && x < b.Width && y >= 0 && y < b.Height
}

// Set writes a cell at (x, y), maintaining wide-rune invariants: a wide
// rune always owns the continuation cell to its right (Rune 0, same
// style), and overwriting either half of a wide pair blanks the other.
func (b *Buffer) Set(x, y int, c Cell) {
	if !b.inBounds(x, y) {
		return
	}
	i := y*b.Width + x

	// Fast path: a narrow rune over a narrow, non-continuation cell —
	// no wide-rune bookkeeping needed. Covers ASCII fills and most
	// writes; the first wide range starts at U+1100.
	if c.Rune < 0x1100 && b.Cells[i].Rune != 0 && b.Cells[i].Rune < 0x1100 {
		b.Cells[i] = c
		return
	}

	// Overwriting the continuation half of a wide rune blanks the rune.
	if b.Cells[i].Rune == 0 && x > 0 && textwidth.Rune(b.Cells[i-1].Rune) == 2 {
		b.Cells[i-1].Rune = ' '
	}
	// Overwriting a wide rune blanks its orphaned continuation cell.
	if textwidth.Rune(b.Cells[i].Rune) == 2 && x+1 < b.Width && b.Cells[i+1].Rune == 0 {
		b.Cells[i+1].Rune = ' '
	}

	if textwidth.Rune(c.Rune) == 2 {
		if x+1 >= b.Width {
			// Wide rune doesn't fit in the last column; paint a blank.
			b.Cells[i] = Cell{Rune: ' ', FG: c.FG, BG: c.BG, Style: c.Style}
			return
		}
		// Claiming the continuation cell may itself split a wide pair.
		if textwidth.Rune(b.Cells[i+1].Rune) == 2 && x+2 < b.Width && b.Cells[i+2].Rune == 0 {
			b.Cells[i+2].Rune = ' '
		}
		b.Cells[i] = c
		b.Cells[i+1] = Cell{Rune: 0, FG: c.FG, BG: c.BG, Style: c.Style}
		return
	}

	b.Cells[i] = c
}

// Get reads a cell at (x, y). Returns empty cell if out of bounds.
func (b *Buffer) Get(x, y int) Cell {
	if !b.inBounds(x, y) {
		return Cell{}
	}
	return b.Cells[y*b.Width+x]
}

// Clear resets all cells to spaces with default colors.
func (b *Buffer) Clear() {
	for i := range b.Cells {
		b.Cells[i] = Cell{Rune: ' '}
	}
}

// WriteString writes a string horizontally starting at (x, y). Wide
// runes occupy two cells; zero-width runes are skipped.
func (b *Buffer) WriteString(x, y int, s string, fg, bg node.Color, style node.StyleFlags) {
	col := x
	for _, r := range s {
		w := textwidth.Rune(r)
		if w == 0 {
			continue
		}
		if col+w > b.Width || !b.inBounds(col, y) {
			break
		}
		b.Set(col, y, Cell{Rune: r, FG: fg, BG: bg, Style: style})
		col += w
	}
}
