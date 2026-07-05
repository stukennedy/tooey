package cell

import (
	"testing"

	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

func TestSetWideRuneClaimsContinuation(t *testing.T) {
	b := NewBuffer(4, 1)
	b.Set(0, 0, Cell{Rune: '日', FG: 1})
	if b.Get(0, 0).Rune != '日' {
		t.Fatal("wide rune not written")
	}
	cont := b.Get(1, 0)
	if cont.Rune != 0 || cont.FG != 1 {
		t.Fatalf("continuation cell not claimed: %+v", cont)
	}
}

func TestSetOverContinuationBlanksWideRune(t *testing.T) {
	b := NewBuffer(4, 1)
	b.Set(0, 0, Cell{Rune: '日'})
	b.Set(1, 0, Cell{Rune: 'X'})
	if b.Get(0, 0).Rune != ' ' {
		t.Fatalf("wide rune should be blanked, got %q", b.Get(0, 0).Rune)
	}
	if b.Get(1, 0).Rune != 'X' {
		t.Fatal("narrow rune not written")
	}
}

func TestSetOverWideRuneBlanksContinuation(t *testing.T) {
	b := NewBuffer(4, 1)
	b.Set(0, 0, Cell{Rune: '日'})
	b.Set(0, 0, Cell{Rune: 'X'})
	if b.Get(1, 0).Rune != ' ' {
		t.Fatalf("continuation should be blanked, got %q", b.Get(1, 0).Rune)
	}
}

func TestSetWideRuneInLastColumn(t *testing.T) {
	b := NewBuffer(2, 1)
	b.Set(1, 0, Cell{Rune: '日', FG: 2})
	got := b.Get(1, 0)
	if got.Rune != ' ' || got.FG != 2 {
		t.Fatalf("wide rune in last column should paint a blank, got %+v", got)
	}
}

func TestSetWideOverWidePair(t *testing.T) {
	b := NewBuffer(4, 1)
	b.Set(1, 0, Cell{Rune: '日'}) // occupies 1,2
	b.Set(0, 0, Cell{Rune: '本'}) // occupies 0,1 — splits 日
	if b.Get(0, 0).Rune != '本' || b.Get(1, 0).Rune != 0 {
		t.Fatal("new wide pair not written")
	}
	if b.Get(2, 0).Rune != ' ' {
		t.Fatalf("orphaned continuation should be blanked, got %q", b.Get(2, 0).Rune)
	}
}

func TestWriteStringWideRunes(t *testing.T) {
	b := NewBuffer(6, 1)
	b.WriteString(0, 0, "a日b", 0, 0, 0)
	if b.Get(0, 0).Rune != 'a' {
		t.Fatal("cell 0 wrong")
	}
	if b.Get(1, 0).Rune != '日' || b.Get(2, 0).Rune != 0 {
		t.Fatal("wide rune should occupy cells 1-2")
	}
	if b.Get(3, 0).Rune != 'b' {
		t.Fatal("narrow rune after wide should land at cell 3")
	}
}

func TestPaintWideText(t *testing.T) {
	n := node.Text("日本")
	lt := layout.Layout(n, 6, 1)
	b := NewBuffer(6, 1)
	Paint(b, lt)
	if b.Get(0, 0).Rune != '日' || b.Get(2, 0).Rune != '本' {
		t.Fatalf("wide text painted incorrectly: %q %q", b.Get(0, 0).Rune, b.Get(2, 0).Rune)
	}
	if b.Get(1, 0).Rune != 0 || b.Get(3, 0).Rune != 0 {
		t.Fatal("continuation cells missing")
	}
}

func TestPaintWideTextClipped(t *testing.T) {
	// 3 columns: 日 fits (0-1), 本 would straddle the edge and must not paint
	n := node.Text("日本")
	lt := layout.Layout(n, 3, 2)
	b := NewBuffer(3, 2)
	Paint(b, lt)
	if b.Get(0, 0).Rune != '日' {
		t.Fatal("first wide rune should paint")
	}
	if b.Get(2, 0).Rune == '本' {
		t.Fatal("straddling wide rune must not paint half out of clip")
	}
}
