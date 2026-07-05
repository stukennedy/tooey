package layout

// HitTest returns the chain of layout nodes containing the point (x, y),
// from the root down to the deepest match. At each level the topmost
// child wins — children later in the list paint over earlier ones, so
// they are checked in reverse order. Returns nil if the point is
// outside the tree.
func HitTest(root LayoutNode, x, y int) []LayoutNode {
	if !contains(root.Rect, x, y) {
		return nil
	}
	path := []LayoutNode{root}
	cur := &root
	for {
		var next *LayoutNode
		for i := len(cur.Children) - 1; i >= 0; i-- {
			if contains(cur.Children[i].Rect, x, y) {
				next = &cur.Children[i]
				break
			}
		}
		if next == nil {
			return path
		}
		path = append(path, *next)
		cur = next
	}
}

func contains(r Rect, x, y int) bool {
	return r.W > 0 && r.H > 0 &&
		x >= r.X && x < r.X+r.W &&
		y >= r.Y && y < r.Y+r.H
}
