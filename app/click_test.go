package app

import (
	"testing"

	"github.com/stukennedy/tooey/focus"
	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

func TestResolveClickOutsideScopeSuppressed(t *testing.T) {
	base := node.Column(
		node.Text("background").WithKey("bg-btn").WithFocusable(),
		node.Text("filler"),
		node.Text("filler"),
	)
	dialog := node.Column(
		node.Text("yes").WithKey("yes").WithFocusable(),
	).WithKey("dlg").WithFocusScope()
	// Modal occupies only the lower rows; the background button stays visible.
	view := node.Overlay(base, node.Column(node.Spacer(), dialog.WithSize(0, 1)))

	lt := layout.Layout(view, 20, 4)
	fm := focus.NewManager()
	fm.Update(lt)
	if fm.ActiveScope() != "dlg" {
		t.Fatalf("expected active scope, got %q", fm.ActiveScope())
	}

	// Click the exposed background button (row 0): suppressed.
	key := resolveClick(layout.HitTest(lt, 1, 0), fm)
	if key != "" {
		t.Fatalf("click outside scope should report no key, got %q", key)
	}
	if fm.Current() != "yes" {
		t.Fatalf("focus must stay in scope, got %q", fm.Current())
	}

	// Click inside the modal (bottom row): reported and focused.
	key = resolveClick(layout.HitTest(lt, 1, 3), fm)
	if key != "yes" {
		t.Fatalf("click inside scope should report its key, got %q", key)
	}
}

func TestResolveClickNoScope(t *testing.T) {
	view := node.Column(node.Text("a").WithKey("a").WithFocusable())
	lt := layout.Layout(view, 10, 2)
	fm := focus.NewManager()
	fm.Update(lt)
	if key := resolveClick(layout.HitTest(lt, 0, 0), fm); key != "a" {
		t.Fatalf("expected 'a', got %q", key)
	}
}
