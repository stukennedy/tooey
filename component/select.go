package component

import (
	"strconv"

	"github.com/stukennedy/tooey/node"
)

// Select renders a single-choice dropdown. Closed, it shows the
// selected option; open, it shows the option list below. Track Open,
// Selected, and HoverIndex in your model and update them from key or
// click messages (each option node is keyed "<Key>-<index>").
type Select struct {
	Key      string
	Options  []string
	Selected int
	// HoverIndex is the option currently highlighted while open.
	HoverIndex int
	Open       bool

	FG          node.Color
	BG          node.Color
	HighlightFG node.Color
	HighlightBG node.Color
}

// Render builds the select. The closed control is keyed with Key and
// focusable, so it can be clicked/tabbed to.
func (s Select) Render() node.Node {
	label := ""
	if s.Selected >= 0 && s.Selected < len(s.Options) {
		label = s.Options[s.Selected]
	}
	arrow := "▾"
	if s.Open {
		arrow = "▴"
	}
	control := node.TextStyled(arrow+" "+label, s.FG, s.BG, 0)
	if s.Key != "" {
		control = control.WithKey(s.Key).WithFocusable()
	}
	if !s.Open {
		return control
	}

	children := make([]node.Node, 0, len(s.Options)+1)
	children = append(children, control)
	for i, opt := range s.Options {
		var n node.Node
		if i == s.HoverIndex {
			n = node.TextStyled("  "+opt, s.HighlightFG, s.HighlightBG, node.Bold)
		} else {
			n = node.TextStyled("  "+opt, s.FG, s.BG, 0)
		}
		if s.Key != "" {
			n = n.WithKey(s.Key + "-" + strconv.Itoa(i)).WithFocusable()
		}
		children = append(children, n)
	}
	return node.Column(children...)
}
