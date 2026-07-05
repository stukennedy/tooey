package layout

import (
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestHitTestFindsDeepestNode(t *testing.T) {
	root := node.Column(
		node.Text("header").WithKey("header"),
		node.Row(
			node.Text("left").WithKey("left").WithFocusable(),
			node.Text("right").WithKey("right").WithFocusable(),
		),
	)
	lt := Layout(root, 20, 5)

	path := HitTest(lt, 1, 0)
	if len(path) == 0 {
		t.Fatal("expected a hit on header row")
	}
	deepest := path[len(path)-1]
	if deepest.Node.Props.Key != "header" {
		t.Fatalf("expected header, got %q", deepest.Node.Props.Key)
	}

	// Row children: "left" is 4 wide starting at x=0, y=1
	path = HitTest(lt, 2, 1)
	deepest = path[len(path)-1]
	if deepest.Node.Props.Key != "left" {
		t.Fatalf("expected left, got %q", deepest.Node.Props.Key)
	}

	path = HitTest(lt, 5, 1)
	deepest = path[len(path)-1]
	if deepest.Node.Props.Key != "right" {
		t.Fatalf("expected right, got %q", deepest.Node.Props.Key)
	}
}

func TestHitTestOutsideTree(t *testing.T) {
	lt := Layout(node.Text("x"), 5, 1)
	if path := HitTest(lt, 50, 50); path != nil {
		t.Fatal("expected nil for out-of-bounds point")
	}
}

func TestHitTestPrefersTopmostOverlayLayer(t *testing.T) {
	base := node.Text("base").WithKey("base")
	layer := node.Text("layer").WithKey("layer")
	lt := Layout(node.Overlay(base, layer), 10, 1)

	path := HitTest(lt, 0, 0)
	deepest := path[len(path)-1]
	if deepest.Node.Props.Key != "layer" {
		t.Fatalf("topmost layer should win, got %q", deepest.Node.Props.Key)
	}
}
