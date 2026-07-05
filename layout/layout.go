package layout

import (
	"strings"

	"github.com/stukennedy/tooey/node"
	"github.com/stukennedy/tooey/textwidth"
)

// Rect is a positioned rectangle in terminal coordinates.
type Rect struct {
	X, Y, W, H int
}

// LayoutNode is a positioned node with resolved layout.
type LayoutNode struct {
	Node     node.Node
	Rect     Rect
	Children []LayoutNode
}

// Layout computes positions for the node tree within the given terminal size.
func Layout(root node.Node, termW, termH int) LayoutNode {
	return layout(root, Rect{0, 0, termW, termH})
}

func layout(n node.Node, avail Rect) LayoutNode {
	ln := LayoutNode{Node: n, Rect: avail}

	switch n.Type {
	case node.TextNode:
		ln = layoutText(n, avail)
	case node.RowNode:
		ln = layoutRow(n, avail)
	case node.ColumnNode, node.ListNode, node.PaneNode:
		ln = layoutColumn(n, avail)
	case node.BoxNode:
		ln = layoutBox(n, avail)
	case node.OverlayNode:
		ln = layoutOverlay(n, avail)
	case node.SpacerNode:
		ln.Rect = avail
	}

	// Apply explicit size constraints
	if n.Props.Width > 0 && n.Props.Width < ln.Rect.W {
		ln.Rect.W = n.Props.Width
	}
	if n.Props.Height > 0 && n.Props.Height < ln.Rect.H {
		ln.Rect.H = n.Props.Height
	}

	return ln
}

func layoutText(n node.Node, avail Rect) LayoutNode {
	pt, pr, pb, pl := padding(n)
	lines := wrapText(n.Props.Text, avail.W-pl-pr)
	h := len(lines) + pt + pb
	if h > avail.H {
		h = avail.H
	}
	// Text uses the full available width (important for flex-allocated space)
	return LayoutNode{
		Node: n,
		Rect: Rect{avail.X, avail.Y, avail.W, h},
	}
}

func layoutRow(n node.Node, avail Rect) LayoutNode {
	ln := LayoutNode{Node: n, Rect: avail}
	if len(n.Children) == 0 {
		return ln
	}
	inner := insetPadding(avail, n)

	// First pass: measure non-flex children
	totalFixed := 0
	totalFlex := 0
	for _, child := range n.Children {
		fw := flexWeight(child)
		if fw > 0 {
			totalFlex += fw
		} else {
			totalFixed += measureWidth(child, inner)
		}
	}

	remaining := inner.W - totalFixed
	if remaining < 0 {
		remaining = 0
	}

	// Second pass: assign positions
	x := inner.X
	for _, child := range n.Children {
		fw := flexWeight(child)
		var childW int
		if fw > 0 && totalFlex > 0 {
			childW = (remaining * fw) / totalFlex
		} else {
			childW = measureWidth(child, inner)
		}
		if childW > inner.W-(x-inner.X) {
			childW = inner.W - (x - inner.X)
		}
		if childW < 0 {
			childW = 0
		}
		childRect := Rect{x, inner.Y, childW, inner.H}
		ln.Children = append(ln.Children, layout(child, childRect))
		x += childW
	}

	return ln
}

func layoutColumn(n node.Node, avail Rect) LayoutNode {
	ln := LayoutNode{Node: n, Rect: avail}
	if len(n.Children) == 0 {
		return ln
	}
	inner := insetPadding(avail, n)

	scrollable := n.Props.ScrollOffset > 0 || n.Props.ScrollToBottom

	// First pass: measure non-flex children
	totalFixed := 0
	totalFlex := 0
	for _, child := range n.Children {
		fw := flexWeight(child)
		if fw > 0 {
			totalFlex += fw
		} else {
			totalFixed += measureHeight(child, inner)
		}
	}

	remaining := inner.H - totalFixed
	if remaining < 0 {
		remaining = 0
	}

	// Second pass: assign positions
	y := inner.Y
	for _, child := range n.Children {
		fw := flexWeight(child)
		var childH int
		if fw > 0 && totalFlex > 0 {
			childH = (remaining * fw) / totalFlex
		} else {
			childH = measureHeight(child, inner)
		}
		if !scrollable {
			if childH > inner.H-(y-inner.Y) {
				childH = inner.H - (y - inner.Y)
			}
			if childH < 0 {
				childH = 0
			}
		}
		childRect := Rect{inner.X, y, inner.W, childH}
		ln.Children = append(ln.Children, layout(child, childRect))
		y += childH
	}

	// Apply scroll offset: shift children upward
	scrollOffset := n.Props.ScrollOffset
	if n.Props.ScrollToBottom {
		totalContentH := y - inner.Y
		if totalContentH > inner.H {
			autoOffset := totalContentH - inner.H
			// Manual scroll (scrollOffset) adjusts from the auto-scroll position
			scrollOffset = autoOffset - n.Props.ScrollOffset
			if scrollOffset < 0 {
				scrollOffset = 0
			}
		} else {
			scrollOffset = 0
		}
	}
	if scrollOffset > 0 {
		for i := range ln.Children {
			shiftY(&ln.Children[i], -scrollOffset)
		}
	}

	return ln
}

func layoutBox(n node.Node, avail Rect) LayoutNode {
	ln := LayoutNode{Node: n, Rect: avail}
	if len(n.Children) == 0 {
		return ln
	}
	// Border takes 1 cell on each side; padding applies inside the border.
	pt, pr, pb, pl := padding(n)
	innerRect := Rect{
		X: avail.X + 1 + pl,
		Y: avail.Y + 1 + pt,
		W: avail.W - 2 - pl - pr,
		H: avail.H - 2 - pt - pb,
	}
	if innerRect.W < 0 {
		innerRect.W = 0
	}
	if innerRect.H < 0 {
		innerRect.H = 0
	}
	ln.Children = append(ln.Children, layout(n.Children[0], innerRect))
	return ln
}

// layoutOverlay stacks every child in the full available rect. Children
// are painted in order, so later children appear on top.
func layoutOverlay(n node.Node, avail Rect) LayoutNode {
	ln := LayoutNode{Node: n, Rect: avail}
	inner := insetPadding(avail, n)
	for _, child := range n.Children {
		ln.Children = append(ln.Children, layout(child, inner))
	}
	return ln
}

// padding returns a node's padding as (top, right, bottom, left).
func padding(n node.Node) (int, int, int, int) {
	return n.Props.PadTop, n.Props.PadRight, n.Props.PadBottom, n.Props.PadLeft
}

// insetPadding shrinks a rect by the node's padding, clamping at zero size.
func insetPadding(r Rect, n node.Node) Rect {
	pt, pr, pb, pl := padding(n)
	r = Rect{X: r.X + pl, Y: r.Y + pt, W: r.W - pl - pr, H: r.H - pt - pb}
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

// measureWidth returns the intrinsic width of a non-flex node.
func measureWidth(n node.Node, avail Rect) int {
	if n.Props.Width > 0 {
		return n.Props.Width
	}
	_, pr, _, pl := padding(n)
	switch n.Type {
	case node.TextNode:
		return textwidth.String(n.Props.Text) + pl + pr
	case node.BoxNode:
		if len(n.Children) > 0 {
			return measureWidth(n.Children[0], avail) + 2 + pl + pr
		}
		return 2 + pl + pr
	case node.RowNode:
		w := 0
		for _, c := range n.Children {
			w += measureWidth(c, avail)
		}
		return w + pl + pr
	case node.OverlayNode:
		// The base layer defines the overlay's intrinsic size.
		if len(n.Children) > 0 {
			return measureWidth(n.Children[0], avail) + pl + pr
		}
		return avail.W
	default:
		return avail.W
	}
}

// measureHeight returns the intrinsic height of a non-flex node.
func measureHeight(n node.Node, avail Rect) int {
	if n.Props.Height > 0 {
		return n.Props.Height
	}
	pt, pr, pb, pl := padding(n)
	switch n.Type {
	case node.TextNode:
		lines := wrapText(n.Props.Text, avail.W-pl-pr)
		return len(lines) + pt + pb
	case node.BoxNode:
		if len(n.Children) > 0 {
			innerAvail := Rect{X: avail.X, Y: avail.Y, W: avail.W - 2 - pl - pr, H: avail.H}
			if innerAvail.W < 0 {
				innerAvail.W = 0
			}
			return measureHeight(n.Children[0], innerAvail) + 2 + pt + pb
		}
		return 2 + pt + pb
	case node.ColumnNode, node.ListNode, node.PaneNode:
		h := 0
		for _, c := range n.Children {
			h += measureHeight(c, avail)
		}
		return h + pt + pb
	case node.RowNode:
		h := 1
		for _, c := range n.Children {
			ch := measureHeight(c, avail)
			if ch > h {
				h = ch
			}
		}
		return h + pt + pb
	case node.OverlayNode:
		// The base layer defines the overlay's intrinsic size.
		if len(n.Children) > 0 {
			return measureHeight(n.Children[0], avail) + pt + pb
		}
		return 1
	default:
		return 1
	}
}

// shiftY recursively shifts a layout node and all descendants by dy.
func shiftY(ln *LayoutNode, dy int) {
	ln.Rect.Y += dy
	for i := range ln.Children {
		shiftY(&ln.Children[i], dy)
	}
}

func flexWeight(n node.Node) int {
	return n.Props.FlexWeight
}

// wrapText wraps text to fit within maxWidth columns.
func wrapText(s string, maxWidth int) []string {
	if maxWidth <= 0 {
		return nil
	}
	if s == "" {
		return []string{""}
	}

	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		// A line that already fits is kept verbatim, preserving
		// internal spacing (e.g. aligned table columns).
		if textwidth.String(paragraph) <= maxWidth {
			lines = append(lines, paragraph)
			continue
		}
		trimmed := strings.TrimLeft(paragraph, " \t")
		leading := paragraph[:len(paragraph)-len(trimmed)]

		if trimmed == "" {
			lines = append(lines, leading)
			continue
		}
		words := strings.Fields(trimmed)
		if len(words) == 0 {
			lines = append(lines, leading)
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
				lines = append(lines, line)
				line = leading + w
				lineLen = textwidth.String(leading) + wLen
			}
		}
		lines = append(lines, line)
	}
	return lines
}
