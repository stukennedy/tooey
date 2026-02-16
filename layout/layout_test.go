package layout

import (
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestTextLayout(t *testing.T) {
	n := node.Text("hello")
	ln := Layout(n, 80, 24)
	if ln.Rect.W != 80 {
		t.Fatalf("expected width 80, got %d", ln.Rect.W)
	}
	if ln.Rect.H != 1 {
		t.Fatalf("expected height 1, got %d", ln.Rect.H)
	}
}

func TestRowFixedChildren(t *testing.T) {
	n := node.Row(
		node.Text("ab"),
		node.Text("cde"),
	)
	ln := Layout(n, 80, 24)
	if len(ln.Children) != 2 {
		t.Fatalf("expected 2 children")
	}
	// measureWidth returns intrinsic text width for allocation
	if ln.Children[0].Rect.X != 0 || ln.Children[0].Rect.W != 2 {
		t.Fatalf("child 0: x=%d w=%d", ln.Children[0].Rect.X, ln.Children[0].Rect.W)
	}
	if ln.Children[1].Rect.X != 2 || ln.Children[1].Rect.W != 3 {
		t.Fatalf("child 1: x=%d w=%d", ln.Children[1].Rect.X, ln.Children[1].Rect.W)
	}
}

func TestRowFlexDistribution(t *testing.T) {
	n := node.Row(
		node.Text("ab"),                   // fixed 2
		node.Spacer(),                     // flex 1
		node.Text("x").WithFlex(2),        // flex 2
	)
	ln := Layout(n, 20, 1)
	// remaining = 20 - 2 = 18, flex total = 3
	// spacer gets 6, flex-2 gets 12
	if ln.Children[1].Rect.W != 6 {
		t.Fatalf("spacer width: expected 6, got %d", ln.Children[1].Rect.W)
	}
	if ln.Children[2].Rect.W != 12 {
		t.Fatalf("flex-2 width: expected 12, got %d", ln.Children[2].Rect.W)
	}
}

func TestColumnStacking(t *testing.T) {
	n := node.Column(
		node.Text("line1"),
		node.Text("line2"),
	)
	ln := Layout(n, 80, 24)
	if ln.Children[0].Rect.Y != 0 {
		t.Fatalf("child 0 y=%d", ln.Children[0].Rect.Y)
	}
	if ln.Children[1].Rect.Y != 1 {
		t.Fatalf("child 1 y=%d", ln.Children[1].Rect.Y)
	}
}

func TestColumnFlexDistribution(t *testing.T) {
	n := node.Column(
		node.Text("top"),
		node.Spacer(),
		node.Text("bottom"),
	)
	ln := Layout(n, 80, 10)
	// top=1, bottom=1, spacer gets 10-2=8
	if ln.Children[1].Rect.H != 8 {
		t.Fatalf("spacer height: expected 8, got %d", ln.Children[1].Rect.H)
	}
}

func TestBoxBorderInset(t *testing.T) {
	n := node.Box(node.BorderSingle, node.Text("hi"))
	ln := Layout(n, 20, 10)
	if len(ln.Children) != 1 {
		t.Fatal("expected 1 child")
	}
	inner := ln.Children[0]
	if inner.Rect.X != 1 || inner.Rect.Y != 1 {
		t.Fatalf("inner pos: (%d,%d)", inner.Rect.X, inner.Rect.Y)
	}
	if inner.Rect.W != 18 {
		t.Fatalf("inner width: expected 18, got %d", inner.Rect.W)
	}
	if inner.Rect.H != 1 {
		t.Fatalf("inner height: expected 1, got %d", inner.Rect.H)
	}
}

func TestTextWrap(t *testing.T) {
	lines := wrapText("hello world foo", 11)
	// "hello world" = 11 <= 11, then "foo" doesn't fit (11+1+3=15>11)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "hello world" {
		t.Fatalf("line 0: %q", lines[0])
	}
	if lines[1] != "foo" {
		t.Fatalf("line 1: %q", lines[1])
	}
}

func TestTextWrapNarrow(t *testing.T) {
	lines := wrapText("hello world foo", 6)
	// "hello" (5<=6), "world" (5<=6), "foo" (3<=6)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

func TestRowWithBoxChildHeight(t *testing.T) {
	// Simulates Indent(2, Box(Column(TextStyled(...)))) — a code block with indent.
	// The Row should get the height of the Box (content + 2 border rows), not 1.
	codeBlock := node.Box(node.BorderRounded, node.Column(
		node.TextStyled("go install github.com/stukennedy/kyotee@latest", 45, 236, 0),
	))
	indented := node.Row(node.Text("  "), codeBlock)

	// Measure height of the Row — should be 3 (1 code line + 2 border rows)
	h := measureHeight(indented, Rect{0, 0, 80, 24})
	if h != 3 {
		t.Fatalf("Row with Box: expected height 3, got %d", h)
	}

	// Full layout: Column containing text, the indented code block, and more text
	tree := node.Column(
		node.Text("To install kyotee, run:"),
		indented,
		node.Text("This downloads the source."),
	)
	ln := Layout(tree, 80, 24)

	// Column should have 3 children
	if len(ln.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(ln.Children))
	}

	// First child: text at Y=0, height=1
	if ln.Children[0].Rect.Y != 0 || ln.Children[0].Rect.H != 1 {
		t.Fatalf("child 0: Y=%d H=%d", ln.Children[0].Rect.Y, ln.Children[0].Rect.H)
	}

	// Second child (indented code block Row): Y=1, height=3
	if ln.Children[1].Rect.Y != 1 {
		t.Fatalf("child 1 (code row): Y=%d, expected 1", ln.Children[1].Rect.Y)
	}
	if ln.Children[1].Rect.H != 3 {
		t.Fatalf("child 1 (code row): H=%d, expected 3", ln.Children[1].Rect.H)
	}

	// Third child: text after code block, Y=4
	if ln.Children[2].Rect.Y != 4 {
		t.Fatalf("child 2 (after code): Y=%d, expected 4", ln.Children[2].Rect.Y)
	}
}

func TestRowMeasureHeightMultiLineBox(t *testing.T) {
	// Box with 3 lines of code: height should be 5 (3 + 2 borders)
	codeBlock := node.Box(node.BorderRounded, node.Column(
		node.Text("line 1"),
		node.Text("line 2"),
		node.Text("line 3"),
	))
	row := node.Row(node.Text("  "), codeBlock)

	h := measureHeight(row, Rect{0, 0, 80, 24})
	if h != 5 {
		t.Fatalf("Row with 3-line Box: expected height 5, got %d", h)
	}
}

func TestExplicitSize(t *testing.T) {
	n := node.Text("hello world that wraps").WithSize(10, 1)
	ln := Layout(n, 80, 24)
	if ln.Rect.W != 10 {
		t.Fatalf("expected width 10, got %d", ln.Rect.W)
	}
}
