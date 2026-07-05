package cell

import (
	"testing"

	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

func TestTextClipsToOwnRectHorizontally(t *testing.T) {
	// An unbreakable word wider than its slot must not bleed into the
	// sibling to its right.
	view := node.Row(
		node.Text("longwordhere").WithSize(5, 1),
		node.Text("BBB"),
	)
	lt := layout.Layout(view, 20, 2)
	b := NewBuffer(20, 2)
	Paint(b, lt)
	if got := row(b, 0)[:8]; got != "longwBBB" {
		t.Fatalf("first row should be clipped text then sibling, got %q", got)
	}
}

func TestTextClipsToOwnRectVertically(t *testing.T) {
	// A height-constrained text must not paint extra lines over the
	// sibling below.
	view := node.Column(
		node.Text("a\nb\nc").WithSize(0, 1),
		node.Text("XXX"),
	)
	lt := layout.Layout(view, 10, 4)
	b := NewBuffer(10, 4)
	Paint(b, lt)
	if b.Get(0, 1).Rune != 'X' {
		t.Fatalf("row 1 should belong to sibling, got %q", b.Get(0, 1).Rune)
	}
	if b.Get(0, 2).Rune == 'c' {
		t.Fatal("clipped text line must not paint below its rect")
	}
}

func TestTabRendersAsSpace(t *testing.T) {
	lt := layout.Layout(node.Text("a\tb"), 10, 1)
	b := NewBuffer(10, 1)
	Paint(b, lt)
	if got := row(b, 0)[:3]; got != "a b" {
		t.Fatalf("tab should render as one space, got %q", got)
	}
}

func TestNoWrapClipsInsteadOfWrapping(t *testing.T) {
	view := node.Text("col1  col2  col3").WithNoWrap()
	lt := layout.Layout(view, 10, 3)
	b := NewBuffer(10, 3)
	Paint(b, lt)
	if got := row(b, 0); got != "col1  col2" {
		t.Fatalf("NoWrap text should clip at edge, got %q", got)
	}
	if b.Get(0, 1).Rune != ' ' {
		t.Fatal("NoWrap text must not wrap onto a second row")
	}
}
