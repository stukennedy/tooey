package node

import "testing"

func TestTextBuilder(t *testing.T) {
	n := Text("hello")
	if n.Type != TextNode {
		t.Fatalf("expected TextNode, got %d", n.Type)
	}
	if n.Props.Text != "hello" {
		t.Fatalf("expected 'hello', got %q", n.Props.Text)
	}
}

func TestRowBuilder(t *testing.T) {
	n := Row(Text("a"), Text("b"))
	if n.Type != RowNode {
		t.Fatalf("expected RowNode")
	}
	if len(n.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(n.Children))
	}
}

func TestChaining(t *testing.T) {
	n := Text("x").WithKey("k1").WithFlex(2).WithFocusable()
	if n.Props.Key != "k1" {
		t.Fatalf("expected key k1")
	}
	if n.Props.FlexWeight != 2 {
		t.Fatalf("expected flex 2")
	}
	if !n.Props.Focusable {
		t.Fatal("expected focusable")
	}
}

func TestSpacer(t *testing.T) {
	n := Spacer()
	if n.Type != SpacerNode {
		t.Fatal("expected SpacerNode")
	}
	if n.Props.FlexWeight != 1 {
		t.Fatal("expected flex weight 1")
	}
}

func TestBox(t *testing.T) {
	n := Box(BorderSingle, Text("content"))
	if n.Type != BoxNode {
		t.Fatal("expected BoxNode")
	}
	if len(n.Children) != 1 {
		t.Fatal("expected 1 child")
	}
	if n.Props.Border != BorderSingle {
		t.Fatal("expected single border")
	}
}

func TestBar(t *testing.T) {
	n := Bar("hello", 1, 2, Bold)
	if n.Props.Text != "hello" {
		t.Fatalf("expected text 'hello', got %q", n.Props.Text)
	}
	if n.Props.FlexWeight != 1 {
		t.Fatalf("expected flex 1, got %d", n.Props.FlexWeight)
	}
	if n.Props.FG != 1 || n.Props.BG != 2 {
		t.Fatal("unexpected colors")
	}
}

func TestSeparator(t *testing.T) {
	n := Separator(5)
	if n.Props.Text != "─────" {
		t.Fatalf("expected 5 dashes, got %q", n.Props.Text)
	}
}

func TestSeparatorStyled(t *testing.T) {
	n := SeparatorStyled('=', 3, 10)
	if n.Props.Text != "===" {
		t.Fatalf("expected '===', got %q", n.Props.Text)
	}
	if n.Props.FG != 10 {
		t.Fatalf("expected FG 10, got %d", n.Props.FG)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		text     string
		max      int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"hello", 0, ""},
	}
	for _, tt := range tests {
		got := Truncate(tt.text, tt.max)
		if got != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.text, tt.max, got, tt.expected)
		}
	}
}

func TestIndent(t *testing.T) {
	n := Indent(4, Text("hi"))
	if n.Type != RowNode {
		t.Fatal("expected RowNode")
	}
	if len(n.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(n.Children))
	}
	if n.Children[0].Props.Text != "    " {
		t.Fatalf("expected 4 spaces, got %q", n.Children[0].Props.Text)
	}
}

func TestPad(t *testing.T) {
	n := Pad(1, 2, 1, 3, Text("x"))
	if n.Type != ColumnNode {
		t.Fatal("expected ColumnNode")
	}
	// top(1) + row + bottom(1) = 3 children
	if len(n.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(n.Children))
	}
}

func TestParagraph(t *testing.T) {
	n := Paragraph("a\nb\nc", 1, 0, 0)
	if n.Type != ColumnNode {
		t.Fatal("expected ColumnNode")
	}
	if len(n.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(n.Children))
	}
	if n.Children[1].Props.Text != "b" {
		t.Fatalf("expected 'b', got %q", n.Children[1].Props.Text)
	}
}

func TestParagraphStyled(t *testing.T) {
	n := ParagraphStyled("line1\nline2", ParagraphOpts{FG: 5, Style: Italic})
	if len(n.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(n.Children))
	}
	if n.Children[0].Props.FG != 5 {
		t.Fatal("expected FG 5")
	}
}
