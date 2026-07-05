package layout

import (
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestColumnPaddingInsetsChildren(t *testing.T) {
	root := node.Column(node.Text("hello")).WithPadding(1, 2, 1, 3)
	lt := Layout(root, 20, 10)
	child := lt.Children[0]
	if child.Rect.X != 3 || child.Rect.Y != 1 {
		t.Fatalf("child should be inset by padding, got %+v", child.Rect)
	}
	if child.Rect.W != 20-3-2 {
		t.Fatalf("child width should be reduced by horizontal padding, got %d", child.Rect.W)
	}
}

func TestRowPaddingInsetsChildren(t *testing.T) {
	root := node.Row(node.Text("ab")).WithPadding(2, 0, 0, 4)
	lt := Layout(root, 20, 10)
	child := lt.Children[0]
	if child.Rect.X != 4 || child.Rect.Y != 2 {
		t.Fatalf("child should be inset by padding, got %+v", child.Rect)
	}
}

func TestBoxPaddingInsideBorder(t *testing.T) {
	root := node.Box(node.BorderSingle, node.Text("x")).WithPaddingAll(1)
	lt := Layout(root, 10, 6)
	child := lt.Children[0]
	if child.Rect.X != 2 || child.Rect.Y != 2 {
		t.Fatalf("box child should be inset by border+padding, got %+v", child.Rect)
	}
	if child.Rect.W != 10-2-2 {
		t.Fatalf("box child width wrong: %d", child.Rect.W)
	}
}

func TestTextPaddingAffectsMeasuredHeight(t *testing.T) {
	col := node.Column(
		node.Text("one").WithPadding(1, 0, 1, 0),
		node.Text("two"),
	)
	lt := Layout(col, 20, 10)
	// Padded text occupies 3 rows, so "two" starts at y=3
	if lt.Children[1].Rect.Y != 3 {
		t.Fatalf("padded text should push sibling down, got y=%d", lt.Children[1].Rect.Y)
	}
}

func TestPadHelperUsesPaddingProps(t *testing.T) {
	root := node.Pad(1, 1, 1, 1, node.Text("x"))
	lt := Layout(root, 10, 5)
	inner := lt.Children[0]
	if inner.Rect.X != 1 || inner.Rect.Y != 1 {
		t.Fatalf("Pad helper should inset child, got %+v", inner.Rect)
	}
}
