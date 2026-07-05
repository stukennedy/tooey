package node

import (
	"strings"

	"github.com/stukennedy/tooey/textwidth"
)

// NodeType identifies the kind of UI node.
type NodeType int

const (
	TextNode NodeType = iota
	BoxNode
	RowNode
	ColumnNode
	ListNode
	PaneNode
	SpacerNode
	OverlayNode
)

// Color represents a terminal color. The zero value is the terminal default.
//
// Encoding (uint32):
//   - 0: terminal default (unset)
//   - 0x000000NN (NN in 1..255): ANSI 256-palette index, so plain literals
//     like Color(245) keep working
//   - ModeAnsi | NN: explicit ANSI 256-palette index — use Ansi(0) for
//     palette black, which the plain form cannot express
//   - ModeRGB | 0xRRGGBB: 24-bit truecolor
type Color uint32

const (
	// ModeRGB marks a Color as 24-bit RGB (low 24 bits are 0xRRGGBB).
	ModeRGB Color = 0x01000000
	// ModeAnsi marks a Color as an explicit ANSI-256 index (low 8 bits).
	ModeAnsi Color = 0x02000000

	colorModeMask Color = 0xFF000000
)

// RGB returns a 24-bit truecolor Color.
func RGB(r, g, b uint8) Color {
	return ModeRGB | Color(r)<<16 | Color(g)<<8 | Color(b)
}

// Ansi returns an explicit ANSI 256-palette Color. Unlike a plain
// Color(n) literal, Ansi(0) means palette black rather than default.
func Ansi(n uint8) Color {
	return ModeAnsi | Color(n)
}

// IsDefault reports whether the color is the terminal default.
func (c Color) IsDefault() bool { return c == 0 }

// IsRGB reports whether the color is a 24-bit RGB color.
func (c Color) IsRGB() bool { return c&colorModeMask == ModeRGB }

// RGBValues returns the red, green, blue components of an RGB color.
func (c Color) RGBValues() (r, g, b uint8) {
	return uint8(c >> 16), uint8(c >> 8), uint8(c)
}

// Ansi256 returns the ANSI 256-palette index for palette colors.
// For RGB colors it returns the nearest palette approximation.
func (c Color) Ansi256() uint8 {
	if c.IsRGB() {
		r, g, b := c.RGBValues()
		return rgbToAnsi256(r, g, b)
	}
	return uint8(c)
}

// rgbToAnsi256 maps 24-bit RGB to the nearest xterm-256 palette entry
// (6x6x6 color cube at 16..231, grayscale ramp at 232..255).
func rgbToAnsi256(r, g, b uint8) uint8 {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return uint8(232 + (int(r)-8)*24/247)
	}
	cube := func(v uint8) int {
		if v < 48 {
			return 0
		}
		if v < 114 {
			return 1
		}
		return int(v-35) / 40
	}
	return uint8(16 + 36*cube(r) + 6*cube(g) + cube(b))
}

// StyleFlags are bitwise text style attributes.
type StyleFlags uint8

const (
	Bold      StyleFlags = 1 << iota
	Dim
	Italic
	Underline
	Reverse
)

// BorderStyle defines box border appearance.
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSingle
	BorderDouble
	BorderRounded
)

// Props holds configurable properties for a node.
type Props struct {
	Text       string
	Width      int // 0 = auto
	Height     int // 0 = auto
	FlexWeight int // 0 = no flex, >0 = relative weight
	Border     BorderStyle
	Focusable  bool
	Key        string
	FG           Color
	BG           Color
	Style        StyleFlags
	ScrollOffset   int  // vertical scroll offset for Column/List/Pane
	ScrollToBottom bool // auto-scroll so bottom content is visible

	// Padding insets content from the node's rect (inside a Box border).
	PadTop, PadRight, PadBottom, PadLeft int

	// NoWrap disables word-wrapping for Text: lines render as-is and
	// clip at the node's edge (used for pre-aligned content like tables).
	NoWrap bool

	// FocusScope traps focus inside this subtree while it is rendered:
	// Tab cycling and click-to-focus only reach focusables within the
	// topmost scope. Opening a scope saves the current focus; removing
	// it restores it. Give scopes a Key so nested scopes stay stable.
	FocusScope bool
}

// Node represents a virtual UI element in the component tree.
type Node struct {
	Type     NodeType
	Props    Props
	Children []Node
}

// Builder functions

func Text(s string) Node {
	return Node{Type: TextNode, Props: Props{Text: s}}
}

func TextStyled(s string, fg, bg Color, style StyleFlags) Node {
	return Node{Type: TextNode, Props: Props{Text: s, FG: fg, BG: bg, Style: style}}
}

func Row(children ...Node) Node {
	return Node{Type: RowNode, Children: children}
}

func Column(children ...Node) Node {
	return Node{Type: ColumnNode, Children: children}
}

func Box(border BorderStyle, child Node) Node {
	return Node{Type: BoxNode, Props: Props{Border: border}, Children: []Node{child}}
}

func List(children ...Node) Node {
	return Node{Type: ListNode, Children: children}
}

func Pane(children ...Node) Node {
	return Node{Type: PaneNode, Children: children}
}

func Spacer() Node {
	return Node{Type: SpacerNode, Props: Props{FlexWeight: 1}}
}

// Overlay stacks children in the same rect: the first child is the base
// layer, later children paint on top of it. Use for modals, popups, and
// dropdowns. Compose with Centered (or Spacers) to position a layer.
func Overlay(children ...Node) Node {
	return Node{Type: OverlayNode, Children: children}
}

// Centered wraps a child so it renders centered in the available space.
// The child keeps its intrinsic size.
func Centered(child Node) Node {
	return Column(
		Spacer(),
		Row(Spacer(), child, Spacer()),
		Spacer(),
	)
}

// WithKey sets the key on a node and returns it.
func (n Node) WithKey(key string) Node {
	n.Props.Key = key
	return n
}

// WithFlex sets the flex weight and returns the node.
func (n Node) WithFlex(weight int) Node {
	n.Props.FlexWeight = weight
	return n
}

// WithSize sets explicit width/height and returns the node.
func (n Node) WithSize(w, h int) Node {
	n.Props.Width = w
	n.Props.Height = h
	return n
}

// WithFocusable marks the node as focusable.
func (n Node) WithFocusable() Node {
	n.Props.Focusable = true
	return n
}

// WithFG sets the foreground color and returns the node.
func (n Node) WithFG(c Color) Node {
	n.Props.FG = c
	return n
}

// WithBG sets the background color and returns the node. On a Box, a
// background fills the interior — useful for modals over an Overlay.
func (n Node) WithBG(c Color) Node {
	n.Props.BG = c
	return n
}

// WithStyle sets the style flags and returns the node.
func (n Node) WithStyle(s StyleFlags) Node {
	n.Props.Style = s
	return n
}

// WithScrollOffset sets the vertical scroll offset.
func (n Node) WithScrollOffset(offset int) Node {
	n.Props.ScrollOffset = offset
	return n
}

func (n Node) WithScrollToBottom() Node {
	n.Props.ScrollToBottom = true
	return n
}

// WithPadding sets per-side padding (top, right, bottom, left) and
// returns the node. Padding insets content from the node's rect; for a
// Box it applies inside the border.
func (n Node) WithPadding(top, right, bottom, left int) Node {
	n.Props.PadTop = top
	n.Props.PadRight = right
	n.Props.PadBottom = bottom
	n.Props.PadLeft = left
	return n
}

// WithPaddingAll sets the same padding on all four sides.
func (n Node) WithPaddingAll(p int) Node {
	return n.WithPadding(p, p, p, p)
}

// WithNoWrap disables word-wrapping; over-wide lines clip at the edge.
func (n Node) WithNoWrap() Node {
	n.Props.NoWrap = true
	return n
}

// WithFocusScope traps focus inside this subtree while it is rendered
// (e.g. a modal overlay layer). See Props.FocusScope.
func (n Node) WithFocusScope() Node {
	n.Props.FocusScope = true
	return n
}

// Bar creates a full-width text node with background color fill.
// Use in a Row; the FlexWeight=1 causes it to stretch to fill available width.
func Bar(text string, fg, bg Color, style StyleFlags) Node {
	return TextStyled(text, fg, bg, style).WithFlex(1)
}

// Separator returns a horizontal line of the given width using "─".
func Separator(width int) Node {
	return TextStyled(strings.Repeat("─", width), 245, 0, 0)
}

// SeparatorStyled returns a horizontal line with a custom character and color.
func SeparatorStyled(ch rune, width int, fg Color) Node {
	return TextStyled(strings.Repeat(string(ch), width), fg, 0, 0)
}

// Truncate truncates text to maxWidth display cells, adding "…" if it
// exceeds the limit. Wide characters count as two cells.
func Truncate(text string, maxWidth int) string {
	return textwidth.Truncate(text, maxWidth)
}

// Indent wraps a child node with left indentation.
func Indent(spaces int, child Node) Node {
	return Row(Text(strings.Repeat(" ", spaces)), child)
}

// Pad wraps a child node with padding on all sides.
func Pad(top, right, bottom, left int, child Node) Node {
	return Column(child).WithPadding(top, right, bottom, left)
}

// ParagraphOpts configures Paragraph rendering.
type ParagraphOpts struct {
	FG    Color
	BG    Color
	Style StyleFlags
}

// Paragraph splits text on newlines and returns a Column of styled text lines.
func Paragraph(text string, fg, bg Color, style StyleFlags) Node {
	return ParagraphStyled(text, ParagraphOpts{FG: fg, BG: bg, Style: style})
}

// ParagraphStyled splits text on newlines using ParagraphOpts.
func ParagraphStyled(text string, opts ParagraphOpts) Node {
	lines := strings.Split(text, "\n")
	children := make([]Node, len(lines))
	for i, line := range lines {
		children[i] = TextStyled(line, opts.FG, opts.BG, opts.Style)
	}
	return Column(children...)
}
