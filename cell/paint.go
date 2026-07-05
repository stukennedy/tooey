package cell

import (
	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
	"github.com/stukennedy/tooey/textwidth"
)

// Paint renders a layout tree into the cell buffer.
func Paint(buf *Buffer, tree layout.LayoutNode) {
	paintNode(buf, tree, tree.Rect)
}

func paintNode(buf *Buffer, ln layout.LayoutNode, clip layout.Rect) {
	r := ln.Rect
	n := ln.Node

	switch n.Type {
	case node.TextNode:
		paintText(buf, ln, clip)
	case node.BoxNode:
		paintBox(buf, n, r, clip)
	}

	// Recurse into children, clipping to the parent's content box
	// (inside padding, and inside a Box border).
	content := r
	if n.Type == node.BoxNode && n.Props.Border != node.BorderNone {
		content = layout.Rect{X: content.X + 1, Y: content.Y + 1, W: content.W - 2, H: content.H - 2}
	}
	content.X += n.Props.PadLeft
	content.Y += n.Props.PadTop
	content.W -= n.Props.PadLeft + n.Props.PadRight
	content.H -= n.Props.PadTop + n.Props.PadBottom
	if content.W < 0 {
		content.W = 0
	}
	if content.H < 0 {
		content.H = 0
	}
	childClip := intersect(content, clip)
	for _, child := range ln.Children {
		paintNode(buf, child, childClip)
	}
}

func paintText(buf *Buffer, ln layout.LayoutNode, clip layout.Rect) {
	n := ln.Node
	r := ln.Rect

	// A text node never paints outside its own rect, even when a line
	// is wider or taller than the space it was given.
	clip = intersect(r, clip)
	if clip.W == 0 || clip.H == 0 {
		return
	}

	// If BG is set, fill the rect so the background shows for spaces.
	if n.Props.BG != 0 {
		fillRect(buf, clip, Cell{Rune: ' ', FG: n.Props.FG, BG: n.Props.BG})
	}

	pt, pl := n.Props.PadTop, n.Props.PadLeft
	lines := ln.Lines
	if lines == nil {
		// Fallback for callers painting a hand-built layout tree.
		lines = textLines(n, r.W-pl-n.Props.PadRight)
	}
	for row, line := range lines {
		y := r.Y + pt + row
		if y < clip.Y {
			continue
		}
		if y >= clip.Y+clip.H {
			break
		}
		col := r.X + pl
		for _, ch := range line {
			w := textwidth.Rune(ch)
			if w == 0 {
				continue
			}
			if col+w > clip.X+clip.W {
				break
			}
			if col >= clip.X {
				buf.Set(col, y, Cell{
					Rune:  ch,
					FG:    n.Props.FG,
					BG:    n.Props.BG,
					Style: n.Props.Style,
				})
			}
			col += w
		}
	}
}

// textLines mirrors layout's wrapping for text painted without a
// layout-computed Lines slice.
func textLines(n node.Node, width int) []string {
	if n.Props.NoWrap {
		return textwidth.SplitLines(n.Props.Text)
	}
	return textwidth.Wrap(n.Props.Text, width)
}

// fillRect writes c to every cell of r (caller pre-clips r).
func fillRect(buf *Buffer, r layout.Rect, c Cell) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			buf.Set(x, y, c)
		}
	}
}

func paintBox(buf *Buffer, n node.Node, r layout.Rect, clip layout.Rect) {
	fg, bg, style := n.Props.FG, n.Props.BG, n.Props.Style

	// A background fills the whole rect (so overlays occlude content
	// beneath them); children paint over it afterwards. Style is left
	// off the fill so underline/reverse don't smear across blank cells.
	if bg != 0 {
		fillRect(buf, intersect(r, clip), Cell{Rune: ' ', FG: fg, BG: bg})
	}

	if r.W < 2 || r.H < 2 {
		return
	}

	var tl, tr, bl, br, hz, vt rune
	switch n.Props.Border {
	case node.BorderSingle:
		tl, tr, bl, br, hz, vt = '┌', '┐', '└', '┘', '─', '│'
	case node.BorderDouble:
		tl, tr, bl, br, hz, vt = '╔', '╗', '╚', '╝', '═', '║'
	case node.BorderRounded:
		tl, tr, bl, br, hz, vt = '╭', '╮', '╰', '╯', '─', '│'
	default:
		return
	}

	setClipped := func(x, y int, ch rune) {
		if x >= clip.X && x < clip.X+clip.W && y >= clip.Y && y < clip.Y+clip.H {
			buf.Set(x, y, Cell{Rune: ch, FG: fg, BG: bg, Style: style})
		}
	}

	// Corners
	setClipped(r.X, r.Y, tl)
	setClipped(r.X+r.W-1, r.Y, tr)
	setClipped(r.X, r.Y+r.H-1, bl)
	setClipped(r.X+r.W-1, r.Y+r.H-1, br)

	// Horizontal edges
	for x := r.X + 1; x < r.X+r.W-1; x++ {
		setClipped(x, r.Y, hz)
		setClipped(x, r.Y+r.H-1, hz)
	}

	// Vertical edges
	for y := r.Y + 1; y < r.Y+r.H-1; y++ {
		setClipped(r.X, y, vt)
		setClipped(r.X+r.W-1, y, vt)
	}
}

func intersect(a, b layout.Rect) layout.Rect {
	x1 := max(a.X, b.X)
	y1 := max(a.Y, b.Y)
	x2 := min(a.X+a.W, b.X+b.W)
	y2 := min(a.Y+a.H, b.Y+b.H)
	if x2 <= x1 || y2 <= y1 {
		return layout.Rect{}
	}
	return layout.Rect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}
}

