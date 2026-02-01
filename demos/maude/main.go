package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stukennedy/tooey/app"
	"github.com/stukennedy/tooey/component"
	"github.com/stukennedy/tooey/input"
	"github.com/stukennedy/tooey/node"

	"golang.org/x/term"
)

// --- Message types ---

type thinkingDoneMsg struct {
	reply chatMessage
}

// --- Data types ---

type role int

const (
	roleUser role = iota
	roleAssistant
)

type toolBlock struct {
	Name    string
	Content string
}

type chatMessage struct {
	Role  role
	Text  string
	Tools []toolBlock
}

type maudeModel struct {
	width, height int
	messages      []chatMessage
	input         component.TextInput
	scrollOffset  int
	thinking      bool
	pendingReply  int
	tokenCount    int
	cost          float64
}

// --- Canned responses ---

var cannedResponses = []chatMessage{
	{
		Role: roleAssistant,
		Text: "I'll help you with that. Let me look at the relevant file first.",
		Tools: []toolBlock{
			{Name: "Read main.go", Content: " 1   package main\n 2\n 3   import \"fmt\"\n 4\n 5   func main() {\n 6       fmt.Println(\"hello\")\n 7   }"},
		},
	},
	{
		Role: roleAssistant,
		Text: "I see the issue. Let me make the changes now.",
		Tools: []toolBlock{
			{Name: "Edit main.go", Content: " 3   import \"fmt\"\n 4\n 5   func main() {\n 6 -     fmt.Println(\"hello\")\n 6 +     fmt.Println(\"hello, world\")\n 7   }"},
		},
	},
	{
		Role: roleAssistant,
		Text: "Let me verify the changes compile correctly.",
		Tools: []toolBlock{
			{Name: "Bash", Content: "$ go build ./...\n\nBuild succeeded."},
		},
	},
	{
		Role: roleAssistant,
		Text: "All done! Here's what I did:\n\n- Read the source file to understand the current code\n- [x] Updated the greeting message\n- [x] Verified the build passes\n- [ ] Run the test suite\n\nSummary:\n  1. Changed \"hello\" to \"hello, world\"\n  2. Build succeeded with no errors",
	},
}

// --- Colors ---

const (
	colWhite    node.Color = 15
	colGray     node.Color = 245
	colDimGray  node.Color = 240
	colDarkGray node.Color = 236
	colBlack    node.Color = 0
	colGreen    node.Color = 2
	colBrGreen  node.Color = 10
	colYellow   node.Color = 3
	colMagenta  node.Color = 5
	colCyan     node.Color = 6
	colBrCyan   node.Color = 14
	colOrange   node.Color = 208
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	a := &app.App{
		Init: func() interface{} {
			w, h := input.TermSize()
			return &maudeModel{
				width:  w,
				height: h,
				input:  component.NewTextInput("Send a message..."),
				messages: []chatMessage{
					{Role: roleAssistant, Text: "Hello! I'm Maude Code, your AI coding assistant. How can I help you today?"},
				},
				tokenCount: 42,
				cost:       0.001,
			}
		},
		Update: maudeUpdate,
		View:   maudeView,
	}

	if err := a.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func maudeUpdate(m interface{}, msg app.Msg) app.UpdateResult {
	mdl := m.(*maudeModel)

	switch msg := msg.(type) {
	case app.ResizeMsg:
		mdl.width, mdl.height = msg.Width, msg.Height

	case app.KeyMsg:
		switch msg.Key.Type {
		case input.Enter:
			if mdl.thinking {
				break
			}
			text, newInput := mdl.input.Submit()
			mdl.input = newInput
			if text == "" {
				return app.NoCmd(mdl)
			}
			mdl.messages = append(mdl.messages, chatMessage{Role: roleUser, Text: text})
			mdl.tokenCount += len(text) / 4
			mdl.thinking = true
			mdl.scrollOffset = 0

			replyIdx := mdl.pendingReply % len(cannedResponses)
			mdl.pendingReply++
			return app.WithCmd(mdl, func() app.Msg {
				time.Sleep(1500 * time.Millisecond)
				return thinkingDoneMsg{reply: cannedResponses[replyIdx]}
			})
		case input.PageUp:
			mdl.scrollOffset += 5
		case input.PageDown:
			mdl.scrollOffset -= 5
			if mdl.scrollOffset < 0 {
				mdl.scrollOffset = 0
			}
		default:
			if !mdl.thinking {
				mdl.input = mdl.input.Update(msg.Key)
			}
		}

	case thinkingDoneMsg:
		mdl.thinking = false
		mdl.messages = append(mdl.messages, msg.reply)
		mdl.tokenCount += len(msg.reply.Text)/4 + 50
		mdl.cost += 0.003
		mdl.scrollOffset = 0

	case app.FocusMsg:
		mdl.input.Focused = msg.Focused

	case app.ScrollMsg:
		mdl.scrollOffset += msg.Delta
		if mdl.scrollOffset < 0 {
			mdl.scrollOffset = 0
		}
	}

	return app.NoCmd(mdl)
}

func maudeView(m interface{}, focused string) node.Node {
	mdl := m.(*maudeModel)
	w := mdl.width

	// --- Status bar ---
	left := " Maude Code"
	right := fmt.Sprintf("opus-4.5 │ tokens: %d │ $%.3f ", mdl.tokenCount, mdl.cost)
	pad := w - len(left) - len(right)
	if pad < 1 {
		pad = 1
	}
	statusText := left + strings.Repeat(" ", pad) + right
	statusBar := node.TextStyled(statusText, colWhite, colDarkGray, node.Bold)

	// --- Separator ---
	sep := node.TextStyled(strings.Repeat("─", w), colDimGray, 0, 0)

	// --- Conversation ---
	var convChildren []node.Node
	for _, msg := range mdl.messages {
		convChildren = append(convChildren, node.Text("")) // blank line spacer
		if msg.Role == roleUser {
			lines := strings.Split(msg.Text, "\n")
			for i, line := range lines {
				prefix := "    "
				if i == 0 {
					prefix = "  > "
				}
				convChildren = append(convChildren,
					node.TextStyled(prefix+line, colBrGreen, 0, node.Bold),
				)
			}
		} else {
			convChildren = append(convChildren, renderAssistantText(msg.Text)...)
			for _, tb := range msg.Tools {
				convChildren = append(convChildren,
					renderToolBlock(tb, w),
				)
			}
		}
	}

	if mdl.thinking {
		convChildren = append(convChildren,
			node.Text(""),
			node.TextStyled("  ● Thinking...", colMagenta, 0, node.Bold),
		)
	}

	conversation := node.Column(convChildren...).
		WithFlex(1).
		WithScrollToBottom().
		WithScrollOffset(mdl.scrollOffset)

	// --- Input area ---
	inputLine := mdl.input.Render("  > ", colWhite, 0, 0)

	// --- Bottom border ---
	bottomBorder := node.TextStyled(strings.Repeat("─", w), colDimGray, 0, 0)

	// --- Help bar ---
	helpText := "Ctrl+C quit"
	helpPad := w - len(helpText)
	if helpPad < 0 {
		helpPad = 0
	}
	helpBar := node.TextStyled(helpText+strings.Repeat(" ", helpPad), colGray, colDarkGray, 0)

	return node.Column(
		statusBar,
		sep,
		conversation,
		sep,
		inputLine,
		bottomBorder,
		helpBar,
	)
}

// Diff background colors (ANSI 256)
const (
	colDiffRedBG     node.Color = 52  // dark red background
	colDiffRedFG     node.Color = 210 // light red text
	colDiffGreenBG   node.Color = 22  // dark green background
	colDiffGreenFG   node.Color = 156 // light green text
	colDiffRedHiBG   node.Color = 88  // brighter red for removed words
	colDiffGreenHiBG node.Color = 28  // brighter green for added words
)

func renderToolBlock(tb toolBlock, maxWidth int) node.Node {
	// Tool name with icon
	var icon string
	var fg node.Color
	isDiff := false
	switch {
	case strings.HasPrefix(tb.Name, "Read"):
		icon = "▸ "
		fg = colCyan
	case strings.HasPrefix(tb.Name, "Edit"):
		icon = "▸ "
		fg = colYellow
		isDiff = true
	case strings.HasPrefix(tb.Name, "Bash"):
		icon = "▸ "
		fg = colOrange
	default:
		icon = "▸ "
		fg = colGray
	}

	title := node.TextStyled("  "+icon+tb.Name, fg, 0, node.Bold)

	var contentNodes []node.Node
	contentNodes = append(contentNodes, title)

	for _, line := range strings.Split(tb.Content, "\n") {
		contentNodes = append(contentNodes, renderContentLine(line, isDiff, maxWidth))
	}

	inner := node.Column(contentNodes...)
	return node.Box(node.BorderRounded, inner)
}

func renderContentLine(line string, isDiff bool, maxWidth int) node.Node {
	pad := "    "

	if !isDiff {
		return node.TextStyled(pad+line, colGray, 0, 0)
	}

	// Detect diff line type by looking for +/- markers after line number
	trimmed := strings.TrimLeft(line, " 0123456789")

	switch {
	case strings.HasPrefix(trimmed, "- "):
		// Removed line — red background, full width
		return renderDiffLine(pad+line, colDiffRedFG, colDiffRedBG, colDiffRedHiBG, maxWidth)
	case strings.HasPrefix(trimmed, "+ "):
		// Added line — green background, full width
		return renderDiffLine(pad+line, colDiffGreenFG, colDiffGreenBG, colDiffGreenHiBG, maxWidth)
	default:
		// Context line
		return node.TextStyled(pad+line, colGray, 0, 0)
	}
}

func renderDiffLine(text string, fg, bg, hiBG node.Color, maxWidth int) node.Node {
	// Pad to fill width for full-line background color
	textLen := len([]rune(text))
	fill := maxWidth - 4 // account for box borders + padding
	if fill > textLen {
		text += strings.Repeat(" ", fill-textLen)
	}
	return node.TextStyled(text, fg, bg, 0)
}

func renderAssistantText(text string) []node.Node {
	var nodes []node.Node
	for _, line := range strings.Split(text, "\n") {
		nodes = append(nodes, renderMarkdownLine(line))
	}
	return nodes
}

func renderMarkdownLine(line string) node.Node {
	// Count leading whitespace for indentation
	trimmed := strings.TrimLeft(line, " ")
	indent := len(line) - len(trimmed)
	pad := "  " + strings.Repeat(" ", indent)

	// Checkbox: - [x] or - [ ]
	if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") {
		content := trimmed[6:]
		return node.Row(
			node.TextStyled(pad+"✔  ", colBrGreen, 0, 0),
			node.TextStyled(content, colWhite, 0, 0),
		)
	}
	if strings.HasPrefix(trimmed, "- [ ] ") {
		content := trimmed[6:]
		return node.Row(
			node.TextStyled(pad+"☐  ", colGray, 0, 0),
			node.TextStyled(content, colGray, 0, node.Dim),
		)
	}

	// Bullet: - text
	if strings.HasPrefix(trimmed, "- ") {
		content := trimmed[2:]
		return node.Row(
			node.TextStyled(pad+"•  ", colCyan, 0, 0),
			node.TextStyled(content, colWhite, 0, 0),
		)
	}

	// Numbered list: 1. text, 2. text, etc.
	if len(trimmed) >= 3 && trimmed[0] >= '0' && trimmed[0] <= '9' {
		dotIdx := strings.Index(trimmed, ". ")
		if dotIdx > 0 && dotIdx <= 3 {
			num := trimmed[:dotIdx+1]
			content := trimmed[dotIdx+2:]
			return node.Row(
				node.TextStyled(pad+num+"  ", colCyan, 0, 0),
				node.TextStyled(content, colWhite, 0, 0),
			)
		}
	}

	// Plain text
	if trimmed == "" {
		return node.Text("")
	}
	return node.TextStyled(pad+trimmed, colWhite, 0, 0)
}
