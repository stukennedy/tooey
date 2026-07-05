package component

import (
	"strings"

	"github.com/stukennedy/tooey/node"
)

// Progress renders a horizontal progress bar. ratio is clamped to
// [0, 1]; width is the bar's total cell width.
func Progress(ratio float64, width int, fg, bg node.Color) node.Node {
	if width <= 0 {
		return node.Text("")
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio*float64(width) + 0.5)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return node.TextStyled(bar, fg, bg, 0)
}
