package focus

import (
	"testing"

	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

func makeTree() layout.LayoutNode {
	tree := node.Column(
		node.Text("a").WithKey("a").WithFocusable(),
		node.Text("b").WithKey("b").WithFocusable(),
		node.Text("c").WithKey("c").WithFocusable(),
	)
	return layout.Layout(tree, 80, 24)
}

func TestTabCycle(t *testing.T) {
	m := NewManager()
	m.Update(makeTree())
	if m.Current() != "a" {
		t.Fatalf("expected 'a', got %q", m.Current())
	}
	m.Next()
	if m.Current() != "b" {
		t.Fatalf("expected 'b', got %q", m.Current())
	}
	m.Next()
	if m.Current() != "c" {
		t.Fatalf("expected 'c', got %q", m.Current())
	}
	m.Next() // wraps
	if m.Current() != "a" {
		t.Fatalf("expected 'a' after wrap, got %q", m.Current())
	}
}

func TestShiftTabCycle(t *testing.T) {
	m := NewManager()
	m.Update(makeTree())
	m.Prev() // wraps to end
	if m.Current() != "c" {
		t.Fatalf("expected 'c', got %q", m.Current())
	}
}

// baseItems is the main UI used by the scope tests.
func baseItems() node.Node {
	return node.Column(
		node.Text("a").WithKey("a").WithFocusable(),
		node.Text("b").WithKey("b").WithFocusable(),
		node.Text("c").WithKey("c").WithFocusable(),
	)
}

func modal(key string, options ...string) node.Node {
	children := make([]node.Node, len(options))
	for i, o := range options {
		children[i] = node.Text(o).WithKey(o).WithFocusable()
	}
	return node.Column(children...).WithKey(key).WithFocusScope()
}

func TestFocusScopeTrapsCycling(t *testing.T) {
	m := NewManager()
	m.Update(layout.Layout(baseItems(), 80, 24))
	m.Next() // on "b"

	// Modal opens: focus jumps into the scope and cycles only there.
	withModal := node.Overlay(baseItems(), modal("dlg", "yes", "no"))
	m.Update(layout.Layout(withModal, 80, 24))

	if m.Current() != "yes" {
		t.Fatalf("expected focus on first scope item, got %q", m.Current())
	}
	if m.FocusableCount() != 2 {
		t.Fatalf("expected 2 focusables in scope, got %d", m.FocusableCount())
	}
	m.Next()
	m.Next() // wraps within the scope
	if m.Current() != "yes" {
		t.Fatalf("cycling should stay inside the scope, got %q", m.Current())
	}
	if m.Focus("a") {
		t.Fatal("focusing outside the scope must be rejected")
	}

	// Modal closes: previous focus is restored.
	m.Update(layout.Layout(baseItems(), 80, 24))
	if m.Current() != "b" {
		t.Fatalf("expected 'b' restored after scope close, got %q", m.Current())
	}
}

func TestFocusScopeNested(t *testing.T) {
	m := NewManager()
	m.Update(layout.Layout(baseItems(), 80, 24))
	m.Next() // "b"

	outer := node.Overlay(baseItems(), modal("outer", "x", "y"))
	m.Update(layout.Layout(outer, 80, 24))
	m.Next() // "y"

	// Inner modal on top of the outer one (topmost scope wins).
	inner := node.Overlay(baseItems(), modal("outer", "x", "y"), modal("inner", "p", "q"))
	m.Update(layout.Layout(inner, 80, 24))
	if m.Current() != "p" {
		t.Fatalf("expected inner scope focus, got %q", m.Current())
	}

	// Close inner: outer focus restored.
	m.Update(layout.Layout(outer, 80, 24))
	if m.Current() != "y" {
		t.Fatalf("expected 'y' restored in outer scope, got %q", m.Current())
	}

	// Close outer: base focus restored.
	m.Update(layout.Layout(baseItems(), 80, 24))
	if m.Current() != "b" {
		t.Fatalf("expected 'b' restored at base, got %q", m.Current())
	}
}

func TestFocusScopeStalePersistsAcrossFrames(t *testing.T) {
	m := NewManager()
	withModal := node.Overlay(baseItems(), modal("dlg", "yes", "no"))
	m.Update(layout.Layout(withModal, 80, 24))
	m.Next() // "no"
	// Same view re-rendered: focus stays put.
	m.Update(layout.Layout(withModal, 80, 24))
	if m.Current() != "no" {
		t.Fatalf("focus should persist across frames, got %q", m.Current())
	}
}

func TestEmptyFocusables(t *testing.T) {
	m := NewManager()
	tree := node.Text("no focus")
	lt := layout.Layout(tree, 80, 24)
	m.Update(lt)
	if m.Current() != "" {
		t.Fatalf("expected empty, got %q", m.Current())
	}
	m.Next() // should not panic
	m.Prev() // should not panic
}
