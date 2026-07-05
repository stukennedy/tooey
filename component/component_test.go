package component

import (
	"strings"
	"testing"

	"github.com/stukennedy/tooey/node"
	"github.com/stukennedy/tooey/textwidth"
	"github.com/stukennedy/tooey/tooeytest"
)

func TestTableAlignsColumns(t *testing.T) {
	tbl := Table{
		Headers: []string{"NAME", "AGE"},
		Rows: [][]string{
			{"alice", "30"},
			{"bob", "7"},
		},
		Selected: -1,
	}
	tooeytest.AssertFrame(t, tbl.Render(), 20, 4, `
		NAME   AGE
		alice  30
		bob    7`)
}

func TestTableWideRuneAlignment(t *testing.T) {
	tbl := Table{
		Headers:  []string{"NAME", "CITY"},
		Rows:     [][]string{{"日本", "Tokyo"}, {"ab", "Berlin"}},
		Selected: -1,
	}
	got := tooeytest.RenderText(tbl.Render(), 20, 4)
	lines := strings.Split(got, "\n")
	// "日本" is display width 4, same as "NAME"; the CITY column must
	// start at the same display column in every row.
	tokyoCol := textwidth.String(lines[1][:strings.Index(lines[1], "Tokyo")])
	berlinCol := textwidth.String(lines[2][:strings.Index(lines[2], "Berlin")])
	if tokyoCol != berlinCol {
		t.Fatalf("columns misaligned with wide runes:\n%s", got)
	}
}

func TestTableSelectionHighlight(t *testing.T) {
	tbl := Table{
		Rows:       [][]string{{"a"}, {"b"}},
		Selected:   1,
		SelectedFG: 0,
		SelectedBG: 6,
	}
	n := tbl.Render()
	if n.Children[1].Props.BG != 6 {
		t.Fatal("selected row should carry highlight colors")
	}
}

func TestTabs(t *testing.T) {
	tabs := Tabs{Labels: []string{"One", "Two"}, Active: 1, ActiveFG: 0, ActiveBG: 6, Key: "tab"}
	n := tabs.Render()
	if len(n.Children) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(n.Children))
	}
	if n.Children[1].Props.BG != 6 || n.Children[1].Props.Style&node.Bold == 0 {
		t.Fatal("active tab should be highlighted")
	}
	if n.Children[0].Props.Key != "tab-One" || !n.Children[0].Props.Focusable {
		t.Fatal("keyed tabs should be focusable with derived keys")
	}
	if got := tooeytest.RenderText(n, 12, 1); got != " One  Two" {
		t.Fatalf("unexpected tab bar: %q", got)
	}
}

func TestProgress(t *testing.T) {
	if got := tooeytest.RenderText(Progress(0.5, 10, 2, 0), 12, 1); got != "█████░░░░░" {
		t.Fatalf("unexpected bar: %q", got)
	}
	if got := tooeytest.RenderText(Progress(-1, 4, 0, 0), 6, 1); got != "░░░░" {
		t.Fatalf("ratio should clamp low: %q", got)
	}
	if got := tooeytest.RenderText(Progress(2, 4, 0, 0), 6, 1); got != "████" {
		t.Fatalf("ratio should clamp high: %q", got)
	}
}

func TestSelectClosed(t *testing.T) {
	s := Select{Key: "sel", Options: []string{"red", "green"}, Selected: 1}
	n := s.Render()
	if n.Type != node.TextNode {
		t.Fatal("closed select should be a single line")
	}
	if got := tooeytest.RenderText(n, 12, 1); got != "▾ green" {
		t.Fatalf("unexpected closed select: %q", got)
	}
	if n.Props.Key != "sel" || !n.Props.Focusable {
		t.Fatal("select control should be keyed and focusable")
	}
}

func TestSelectOpen(t *testing.T) {
	s := Select{Key: "sel", Options: []string{"red", "green"}, Selected: 0, HoverIndex: 1, Open: true}
	n := s.Render()
	if len(n.Children) != 3 {
		t.Fatalf("open select should show control + options, got %d children", len(n.Children))
	}
	if n.Children[2].Props.Key != "sel-1" {
		t.Fatalf("option keys should be indexed, got %q", n.Children[2].Props.Key)
	}
	if n.Children[2].Props.Style&node.Bold == 0 {
		t.Fatal("hovered option should be highlighted")
	}
}
