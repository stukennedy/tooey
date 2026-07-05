package wire

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestRoundTrip(t *testing.T) {
	root := node.Column(
		node.TextStyled("title", node.RGB(255, 128, 0), node.Ansi(0), node.Bold|node.Underline),
		node.Box(node.BorderRounded, node.Text("body")).WithPaddingAll(1).WithBG(4),
		node.Row(
			node.Text("ok").WithKey("btn-ok").WithFocusable(),
			node.Spacer(),
		),
		node.Overlay(node.Text("base"), node.Text("layer")),
	).WithFlex(1).WithScrollToBottom()

	data, err := Marshal(root)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(root, got) {
		t.Fatalf("round-trip mismatch:\nwant %+v\ngot  %+v", root, got)
	}
}

func TestWireFormatShape(t *testing.T) {
	data, err := Marshal(node.TextStyled("hi", node.RGB(255, 0, 0), 0, node.Bold))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{`"type":"text"`, `"text":"hi"`, `"fg":"#ff0000"`, `"style":["bold"]`} {
		if !strings.Contains(s, want) {
			t.Errorf("wire JSON missing %s: %s", want, s)
		}
	}
}

func TestPlainTextNodeOmitsProps(t *testing.T) {
	data, err := Marshal(node.Row(node.Node{Type: node.TextNode}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "props") {
		t.Fatalf("empty props should be omitted: %s", data)
	}
}

func TestUnknownTypeErrors(t *testing.T) {
	if _, err := Unmarshal([]byte(`{"type":"blink"}`)); err == nil {
		t.Fatal("expected error for unknown node type")
	}
}

func TestUnknownStyleErrors(t *testing.T) {
	if _, err := Unmarshal([]byte(`{"type":"text","props":{"style":["sparkle"]}}`)); err == nil {
		t.Fatal("expected error for unknown style")
	}
}

func TestColorPaletteZeroMeansBlack(t *testing.T) {
	n, err := Unmarshal([]byte(`{"type":"text","props":{"fg":0}}`))
	if err != nil {
		t.Fatal(err)
	}
	if n.Props.FG != node.Ansi(0) {
		t.Fatalf("fg 0 should decode to explicit palette black, got %v", n.Props.FG)
	}
}

func TestActionJSON(t *testing.T) {
	a := Action{Name: "submit", Key: "input-1", Value: "hello"}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var got Action
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "submit" || got.Key != "input-1" || got.Value != "hello" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}
