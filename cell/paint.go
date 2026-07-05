package cell

import (
	"strings"

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
		paintText(buf, n, r, clip)
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

func paintText(buf *Buffer, n node.Node, r layout.Rect, clip layout.Rect) {
	// First, if BG is set, fill the rect so background shows for spaces
	if n.Props.BG != 0 {
		for y := r.Y; y < r.Y+r.H && y < clip.Y+clip.H; y++ {
			if y < clip.Y {
				continue
			}
			for x := r.X; x < r.X+r.W && x < clip.X+clip.W; x++ {
				if x >= clip.X {
					buf.Set(x, y, Cell{Rune: ' ', BG: n.Props.BG})
				}
			}
		}
	}

	pt, pr, _, pl := n.Props.PadTop, n.Props.PadRight, n.Props.PadBottom, n.Props.PadLeft
	lines := wrapText(n.Props.Text, r.W-pl-pr)
	for row, line := range lines {
		y := r.Y + pt + row
		if y < clip.Y || y >= clip.Y+clip.H {
			continue
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

func paintBox(buf *Buffer, n node.Node, r layout.Rect, clip layout.Rect) {
	fg, bg, style := n.Props.FG, n.Props.BG, n.Props.Style

	// A background fills the whole rect (so overlays occlude content
	// beneath them); children paint over it afterwards.
	if bg != 0 {
		vis := intersect(r, clip)
		for y := vis.Y; y < vis.Y+vis.H; y++ {
			for x := vis.X; x < vis.X+vis.W; x++ {
				buf.Set(x, y, Cell{Rune: ' ', FG: fg, BG: bg, Style: style})
			}
		}
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

// wrapText splits text into lines that fit within maxWidth, preserving leading whitespace.
func wrapText(s string, maxWidth int) []string {
	if maxWidth <= 0 {
		return nil
	}
	if s == "" {
		return []string{""}
	}

	// Split on existing newlines first
	rawLines := strings.Split(s, "\n")
	var result []string
	for _, raw := range rawLines {
		// A line that already fits is kept verbatim, preserving
		// internal spacing (e.g. aligned table columns).
		if textwidth.String(raw) <= maxWidth {
			result = append(result, raw)
			continue
		}
		// Preserve leading whitespace
		trimmed := strings.TrimLeft(raw, " \t")
		leading := raw[:len(raw)-len(trimmed)]

		if trimmed == "" {
			result = append(result, leading)
			continue
		}

		words := strings.Fields(trimmed)
		if len(words) == 0 {
			result = append(result, leading)
			continue
		}

		line := leading + words[0]
		lineLen := textwidth.String(line)
		for _, w := range words[1:] {
			wLen := textwidth.String(w)
			if lineLen+1+wLen <= maxWidth {
				line += " " + w
				lineLen += 1 + wLen
			} else {
				result = append(result, line)
				line = leading + w
				lineLen = textwidth.String(leading) + wLen
			}
		}
		result = append(result, line)
	}
	return result
}
