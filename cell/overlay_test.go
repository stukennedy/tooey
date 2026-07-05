package cell

import (
	"testing"

	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

// row extracts the runes of row y as a string, mapping continuation
// cells to spaces.
func row(b *Buffer, y int) string {
	out := make([]rune, b.Width)
	for x := 0; x < b.Width; x++ {
		r := b.Get(x, y).Rune
		if r == 0 {
			r = ' '
		}
		out[x] = r
	}
	return string(out)
}

func TestOverlayPaintsLayersInOrder(t *testing.T) {
	base := node.Text("AAAAAAAAAA")
	layer := node.Text("BB")
	root := node.Overlay(base, layer)

	lt := layout.Layout(root, 10, 1)
	b := NewBuffer(10, 1)
	Paint(b, lt)

	if got := row(b, 0); got != "BBAAAAAAAA" {
		t.Fatalf("layer should paint over base, got %q", got)
	}
}

func TestOverlayCenteredModal(t *testing.T) {
	base := node.Column(
		node.Text("XXXXXXXXXXXXXXXXXXXX"),
		node.Text("XXXXXXXXXXXXXXXXXXXX"),
		node.Text("XXXXXXXXXXXXXXXXXXXX"),
		node.Text("XXXXXXXXXXXXXXXXXXXX"),
		node.Text("XXXXXXXXXXXXXXXXXXXX"),
	)
	modal := node.Box(node.BorderSingle, node.Text("hi")).WithSize(6, 3).WithBG(4)
	root := node.Overlay(base, node.Centered(modal))

	lt := layout.Layout(root, 20, 5)
	b := NewBuffer(20, 5)
	Paint(b, lt)

	// Modal box (6 wide, 3 tall) centers at x=7, y=1
	if b.Get(7, 1).Rune != '┌' {
		t.Fatalf("expected modal corner at (7,1), got %q", b.Get(7, 1).Rune)
	}
	// Interior is filled by the box BG, occluding the base
	inside := b.Get(8, 1)
	if inside.Rune == 'X' {
		t.Fatal("modal interior should occlude base content")
	}
	// Base still visible outside the modal
	if b.Get(0, 0).Rune != 'X' {
		t.Fatal("base should be visible outside the modal")
	}
}

func TestBoxBGFillsInterior(t *testing.T) {
	root := node.Box(node.BorderSingle, node.Text("")).WithBG(2)
	lt := layout.Layout(root, 4, 3)
	b := NewBuffer(4, 3)
	Paint(b, lt)
	c := b.Get(1, 1)
	if c.BG != 2 {
		t.Fatalf("box interior should carry BG, got %+v", c)
	}
}
