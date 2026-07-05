package layout

import (
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestMeasureMultilineTextUsesWidestLine(t *testing.T) {
	view := Layout(node.Row(node.Text("aa\nbbbb"), node.Text("Z")), 20, 3)
	if view.Children[0].Rect.W != 4 {
		t.Fatalf("multi-line text width should be widest line (4), got %d", view.Children[0].Rect.W)
	}
	if view.Children[1].Rect.X != 4 {
		t.Fatalf("sibling should start after widest line, got x=%d", view.Children[1].Rect.X)
	}
}

func TestBorderNoneBoxDoesNotInset(t *testing.T) {
	lt := Layout(node.Box(node.BorderNone, node.Text("hi")), 10, 3)
	child := lt.Children[0]
	if child.Rect.X != 0 || child.Rect.Y != 0 {
		t.Fatalf("borderless box must not reserve border cells, got %+v", child.Rect)
	}
	// And a bordered box still does.
	lt = Layout(node.Box(node.BorderSingle, node.Text("hi")), 10, 3)
	if lt.Children[0].Rect.X != 1 {
		t.Fatalf("bordered box should inset by 1, got %+v", lt.Children[0].Rect)
	}
}

func TestOverlayMeasuredByLargestLayer(t *testing.T) {
	ov := node.Overlay(node.Text("x"), node.Column(node.Text("one"), node.Text("two")))
	view := Layout(node.Column(ov, node.Text("below")), 30, 5)
	if view.Children[0].Rect.H != 2 {
		t.Fatalf("overlay height should fit tallest layer (2), got %d", view.Children[0].Rect.H)
	}
	if view.Children[1].Rect.Y != 2 {
		t.Fatalf("sibling should sit below tallest layer, got y=%d", view.Children[1].Rect.Y)
	}
}

func TestTextLayoutCachesLines(t *testing.T) {
	lt := Layout(node.Text("hello world"), 5, 5)
	if len(lt.Lines) != 2 {
		t.Fatalf("layout should cache wrapped lines, got %v", lt.Lines)
	}
}
