// Package wire defines a JSON serialization for tooey node trees and
// client actions, enabling server-driven UIs: a server (in any
// language) sends node trees over SSE or HTTP, and a thin tooey client
// renders them and posts actions back.
//
// The format is deliberately language-neutral: node types and borders
// are strings, style flags are arrays of strings, and colors are either
// palette integers or "#RRGGBB" strings.
package wire

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/stukennedy/tooey/node"
)

// Node is the wire representation of a node.Node.
type Node struct {
	Type     string `json:"type"`
	Props    *Props `json:"props,omitempty"`
	Children []Node `json:"children,omitempty"`
}

// Props is the wire representation of node.Props. Zero-valued fields
// are omitted from the JSON.
type Props struct {
	Text           string  `json:"text,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	Flex           int     `json:"flex,omitempty"`
	Border         string  `json:"border,omitempty"`
	Focusable      bool    `json:"focusable,omitempty"`
	Key            string  `json:"key,omitempty"`
	FG             *Color  `json:"fg,omitempty"`
	BG             *Color  `json:"bg,omitempty"`
	Style          []string `json:"style,omitempty"`
	ScrollOffset   int     `json:"scrollOffset,omitempty"`
	ScrollToBottom bool    `json:"scrollToBottom,omitempty"`
	Padding        *[4]int `json:"padding,omitempty"` // top, right, bottom, left
	NoWrap         bool    `json:"noWrap,omitempty"`
}

// Color wraps node.Color with a wire-friendly JSON form: a palette
// index (integer, 0 = palette black) or a "#RRGGBB" string. An absent
// field means the terminal default.
type Color struct {
	node.Color
}

// MarshalJSON encodes RGB colors as "#RRGGBB" and palette colors as
// integers.
func (c Color) MarshalJSON() ([]byte, error) {
	if c.IsRGB() {
		r, g, b := c.RGBValues()
		return json.Marshal(fmt.Sprintf("#%02x%02x%02x", r, g, b))
	}
	return json.Marshal(c.Ansi256())
}

// UnmarshalJSON accepts a palette integer or a "#RRGGBB" string.
func (c *Color) UnmarshalJSON(data []byte) error {
	var n uint8
	if err := json.Unmarshal(data, &n); err == nil {
		if n == 0 {
			c.Color = node.Ansi(0) // explicit 0 means palette black
		} else {
			c.Color = node.Color(n)
		}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("wire: color must be a palette integer or \"#RRGGBB\" string: %w", err)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return fmt.Errorf("wire: invalid color %q", s)
	}
	c.Color = node.RGB(r, g, b)
	return nil
}

// Action is a client→server message reporting user interaction, e.g. a
// click on a keyed node or a submitted input value.
type Action struct {
	// Name identifies the kind of action, e.g. "click", "submit".
	Name string `json:"name"`
	// Key is the node key the action originated from, if any.
	Key string `json:"key,omitempty"`
	// Value carries the action payload, e.g. an input's text.
	Value string `json:"value,omitempty"`
	// Data carries arbitrary structured payload.
	Data map[string]any `json:"data,omitempty"`
}

var typeNames = map[node.NodeType]string{
	node.TextNode:    "text",
	node.BoxNode:     "box",
	node.RowNode:     "row",
	node.ColumnNode:  "column",
	node.ListNode:    "list",
	node.PaneNode:    "pane",
	node.SpacerNode:  "spacer",
	node.OverlayNode: "overlay",
}

var typeValues = invert(typeNames)

var borderNames = map[node.BorderStyle]string{
	node.BorderNone:    "",
	node.BorderSingle:  "single",
	node.BorderDouble:  "double",
	node.BorderRounded: "rounded",
}

var borderValues = invert(borderNames)

var styleNames = []struct {
	flag node.StyleFlags
	name string
}{
	{node.Bold, "bold"},
	{node.Dim, "dim"},
	{node.Italic, "italic"},
	{node.Underline, "underline"},
	{node.Reverse, "reverse"},
}

func invert[K comparable, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// FromNode converts a node.Node tree to its wire representation.
func FromNode(n node.Node) Node {
	wn := Node{Type: typeNames[n.Type]}
	if p := fromProps(n.Props); p != nil {
		wn.Props = p
	}
	for _, c := range n.Children {
		wn.Children = append(wn.Children, FromNode(c))
	}
	return wn
}

func fromProps(p node.Props) *Props {
	wp := Props{
		Text:           p.Text,
		Width:          p.Width,
		Height:         p.Height,
		Flex:           p.FlexWeight,
		Border:         borderNames[p.Border],
		Focusable:      p.Focusable,
		Key:            p.Key,
		ScrollOffset:   p.ScrollOffset,
		ScrollToBottom: p.ScrollToBottom,
		NoWrap:         p.NoWrap,
	}
	if !p.FG.IsDefault() {
		wp.FG = &Color{p.FG}
	}
	if !p.BG.IsDefault() {
		wp.BG = &Color{p.BG}
	}
	for _, s := range styleNames {
		if p.Style&s.flag != 0 {
			wp.Style = append(wp.Style, s.name)
		}
	}
	if p.PadTop != 0 || p.PadRight != 0 || p.PadBottom != 0 || p.PadLeft != 0 {
		wp.Padding = &[4]int{p.PadTop, p.PadRight, p.PadBottom, p.PadLeft}
	}
	// Omit fully-default props; DeepEqual keeps this correct as fields
	// are added without hand-maintaining a field list.
	if reflect.DeepEqual(wp, Props{}) {
		return nil
	}
	return &wp
}

// ToNode converts a wire node back to a node.Node tree. Unknown node
// types, borders, or styles are an error.
func (wn Node) ToNode() (node.Node, error) {
	t, ok := typeValues[wn.Type]
	if !ok {
		return node.Node{}, fmt.Errorf("wire: unknown node type %q", wn.Type)
	}
	n := node.Node{Type: t}
	if wn.Props != nil {
		p, err := wn.Props.toProps()
		if err != nil {
			return node.Node{}, err
		}
		n.Props = p
	}
	for _, c := range wn.Children {
		child, err := c.ToNode()
		if err != nil {
			return node.Node{}, err
		}
		n.Children = append(n.Children, child)
	}
	return n, nil
}

func (wp Props) toProps() (node.Props, error) {
	border, ok := borderValues[wp.Border]
	if !ok {
		return node.Props{}, fmt.Errorf("wire: unknown border %q", wp.Border)
	}
	p := node.Props{
		Text:           wp.Text,
		Width:          wp.Width,
		Height:         wp.Height,
		FlexWeight:     wp.Flex,
		Border:         border,
		Focusable:      wp.Focusable,
		Key:            wp.Key,
		ScrollOffset:   wp.ScrollOffset,
		ScrollToBottom: wp.ScrollToBottom,
		NoWrap:         wp.NoWrap,
	}
	if wp.FG != nil {
		p.FG = wp.FG.Color
	}
	if wp.BG != nil {
		p.BG = wp.BG.Color
	}
	for _, name := range wp.Style {
		found := false
		for _, s := range styleNames {
			if s.name == name {
				p.Style |= s.flag
				found = true
				break
			}
		}
		if !found {
			return node.Props{}, fmt.Errorf("wire: unknown style %q", name)
		}
	}
	if wp.Padding != nil {
		p.PadTop, p.PadRight, p.PadBottom, p.PadLeft = wp.Padding[0], wp.Padding[1], wp.Padding[2], wp.Padding[3]
	}
	return p, nil
}

// Marshal encodes a node tree as JSON.
func Marshal(n node.Node) ([]byte, error) {
	return json.Marshal(FromNode(n))
}

// Unmarshal decodes a JSON node tree.
func Unmarshal(data []byte) (node.Node, error) {
	var wn Node
	if err := json.Unmarshal(data, &wn); err != nil {
		return node.Node{}, err
	}
	return wn.ToNode()
}
