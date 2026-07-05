package component

import (
	"strings"

	"github.com/stukennedy/tooey/node"
	"github.com/stukennedy/tooey/textwidth"
)

// Table renders rows of cells in aligned columns, with an optional
// header row and selection highlight.
type Table struct {
	Headers []string
	Rows    [][]string

	// Selected highlights a row when >= 0.
	Selected int

	HeaderFG   node.Color
	HeaderBG   node.Color
	FG         node.Color
	BG         node.Color
	SelectedFG node.Color
	SelectedBG node.Color

	// ColGap is the number of spaces between columns (default 2).
	ColGap int
}

// Render builds the table as a Column of text rows.
func (t Table) Render() node.Node {
	gap := t.ColGap
	if gap <= 0 {
		gap = 2
	}
	widths := t.columnWidths()

	children := make([]node.Node, 0, len(t.Rows)+1)
	if len(t.Headers) > 0 {
		children = append(children,
			node.TextStyled(joinRow(t.Headers, widths, gap), t.HeaderFG, t.HeaderBG, node.Bold))
	}
	for i, row := range t.Rows {
		fg, bg, style := t.FG, t.BG, node.StyleFlags(0)
		if i == t.Selected {
			fg, bg, style = t.SelectedFG, t.SelectedBG, node.Bold
		}
		children = append(children, node.TextStyled(joinRow(row, widths, gap), fg, bg, style))
	}
	return node.Column(children...)
}

func (t Table) columnWidths() []int {
	cols := len(t.Headers)
	for _, r := range t.Rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	widths := make([]int, cols)
	for i, h := range t.Headers {
		widths[i] = textwidth.String(h)
	}
	for _, r := range t.Rows {
		for i, c := range r {
			if w := textwidth.String(c); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func joinRow(cells []string, widths []int, gap int) string {
	var b strings.Builder
	for i, c := range cells {
		b.WriteString(c)
		if i < len(cells)-1 {
			pad := widths[i] - textwidth.String(c) + gap
			b.WriteString(strings.Repeat(" ", pad))
		}
	}
	return b.String()
}
