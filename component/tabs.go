package component

import "github.com/stukennedy/tooey/node"

// Tabs renders a horizontal tab bar with one active tab.
type Tabs struct {
	Labels []string
	Active int

	FG       node.Color
	ActiveFG node.Color
	ActiveBG node.Color

	// Key prefixes each tab's node key ("<Key>-<label>") and makes the
	// tabs focusable/clickable when set.
	Key string
}

// Render builds the tab bar as a Row of labels.
func (t Tabs) Render() node.Node {
	children := make([]node.Node, 0, len(t.Labels))
	for i, label := range t.Labels {
		var n node.Node
		if i == t.Active {
			n = node.TextStyled(" "+label+" ", t.ActiveFG, t.ActiveBG, node.Bold)
		} else {
			n = node.TextStyled(" "+label+" ", t.FG, 0, 0)
		}
		if t.Key != "" {
			n = n.WithKey(t.Key + "-" + label).WithFocusable()
		}
		children = append(children, n)
	}
	return node.Row(children...)
}
