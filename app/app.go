package app

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/stukennedy/tooey/ansi"
	"github.com/stukennedy/tooey/cell"
	"github.com/stukennedy/tooey/diff"
	"github.com/stukennedy/tooey/focus"
	"github.com/stukennedy/tooey/input"
	"github.com/stukennedy/tooey/layout"
	"github.com/stukennedy/tooey/node"
)

// Msg is any message that can trigger a state update.
type Msg interface{}

// KeyMsg wraps a key event as a message.
type KeyMsg struct {
	Key input.Key
}

// ResizeMsg wraps a resize event.
type ResizeMsg struct {
	Width, Height int
}

// FocusMsg indicates the terminal gained or lost focus.
type FocusMsg struct {
	Focused bool
}

// FocusChangedMsg reports that the focused node changed (Tab, click, or
// a focus scope opening/closing). Key is the newly focused node's key,
// "" if nothing is focused. Mirror it into your model when Update needs
// to know what is focused (e.g. Enter activating the focused button).
type FocusChangedMsg struct {
	Key string
}

// DismissMsg reports Escape pressed while a focus scope was active.
// Scope is the scope's identity key (Props.Key when set). Handle it by
// removing the scope from your view (e.g. closing the modal); Escape
// arrives as a plain KeyMsg only when no scope is active.
type DismissMsg struct {
	Scope string
}

// PasteMsg carries text from a bracketed paste event.
type PasteMsg struct {
	Text string
}

// ScrollMsg indicates a mouse scroll event. Delta is positive for scroll
// up, negative for scroll down. X, Y is the cell under the cursor.
type ScrollMsg struct {
	Delta int
	X, Y  int
}

// ClickMsg indicates a mouse click. Key is the key of the deepest keyed
// node under the cursor ("" if none). Clicking a focusable node also
// moves focus to it before Update runs. While a focus scope (modal) is
// active, clicks outside the scope report Key "" — background controls
// cannot be activated or focused through a modal, but the ClickMsg
// still arrives so apps can e.g. dismiss on outside click.
type ClickMsg struct {
	X, Y int
	Key  string
}

// Cmd is a function that runs asynchronously and returns a Msg.
type Cmd func() Msg

// Sub is a long-running command that can send multiple messages via the send callback.
// It returns a final Msg when done (or nil).
type Sub func(send func(Msg)) Msg

// UpdateResult is returned from Update: new model + optional async commands.
// Set Quit (via the Quit helper) to stop the application.
type UpdateResult[M any] struct {
	Model M
	Quit  bool
	Cmds  []Cmd
	Subs  []Sub
}

// NoCmd returns an UpdateResult with no commands.
func NoCmd[M any](model M) UpdateResult[M] {
	return UpdateResult[M]{Model: model}
}

// WithCmd returns an UpdateResult with commands to execute.
func WithCmd[M any](model M, cmds ...Cmd) UpdateResult[M] {
	return UpdateResult[M]{Model: model, Cmds: cmds}
}

// WithSub returns an UpdateResult with subscriptions.
func WithSub[M any](model M, subs ...Sub) UpdateResult[M] {
	return UpdateResult[M]{Model: model, Subs: subs}
}

// Quit returns an UpdateResult that stops the application.
func Quit[M any](model M) UpdateResult[M] {
	return UpdateResult[M]{Model: model, Quit: true}
}

// App defines an Elm-style TUI application over a model type M.
type App[M any] struct {
	// Init returns the initial model.
	Init func() M

	// Update processes a message and returns the new model + optional commands.
	// Return Quit(model) to stop the application.
	Update func(model M, msg Msg) UpdateResult[M]

	// View renders the model to a node tree.
	View func(model M, focused string) node.Node

	// Output writer (defaults to os.Stdout).
	Output io.Writer

	// Input reader (defaults to os.Stdin).
	Input io.Reader
}

// resolveClick converts a click hit path into the key to report,
// moving focus to the clicked focusable. While a focus scope is active,
// clicks that land outside the scope's subtree report no key and cannot
// move focus.
func resolveClick(path []layout.LayoutNode, fm *focus.Manager) string {
	if fm.ActiveScope() != "" {
		inScope := false
		for _, ln := range path {
			if ln.Node.Props.FocusScope {
				inScope = true
				break
			}
		}
		if !inScope {
			return ""
		}
	}
	key := ""
	for i := len(path) - 1; i >= 0; i-- {
		props := path[i].Node.Props
		if key == "" && props.Key != "" {
			key = props.Key
		}
		if props.Focusable && props.Key != "" {
			fm.Focus(props.Key)
			break
		}
	}
	return key
}

// Run starts the application main loop.
func (a *App[M]) Run(ctx context.Context) error {
	out := a.Output
	if out == nil {
		out = os.Stdout
	}
	in := a.Input
	if in == nil {
		in = os.Stdin
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Terminal setup
	ansi.EnterAltScreen(out)
	ansi.HideCursor(out)
	ansi.EnableFocusReporting(out)
	ansi.EnableMouseReporting(out)
	ansi.EnableBracketedPaste(out)
	ansi.ClearScreen(out)
	defer func() {
		ansi.DisableBracketedPaste(out)
		ansi.DisableMouseReporting(out)
		ansi.DisableFocusReporting(out)
		ansi.ShowCursor(out)
		ansi.LeaveAltScreen(out)
	}()

	// Get terminal size
	width, height := input.TermSize()

	model := a.Init()
	fm := focus.NewManager()

	var prevBuf *cell.Buffer
	var lastLayout *layout.LayoutNode
	var lastFocused string

	// toMsg translates a raw key event into an app message; nil means
	// the event is consumed (e.g. mouse button release).
	toMsg := func(k input.Key) Msg {
		switch k.Type {
		case input.FocusIn:
			return FocusMsg{Focused: true}
		case input.FocusOut:
			return FocusMsg{Focused: false}
		case input.MouseScrollUp:
			return ScrollMsg{Delta: 3, X: k.MouseX, Y: k.MouseY}
		case input.MouseScrollDown:
			return ScrollMsg{Delta: -3, X: k.MouseX, Y: k.MouseY}
		case input.MouseRelease:
			return nil
		case input.MouseClick:
			key := ""
			if lastLayout != nil {
				key = resolveClick(layout.HitTest(*lastLayout, k.MouseX, k.MouseY), fm)
			}
			return ClickMsg{X: k.MouseX, Y: k.MouseY, Key: key}
		case input.Escape:
			// Escape dismisses the active focus scope (modal) rather
			// than arriving as a raw key.
			if s := fm.ActiveScope(); s != "" {
				return DismissMsg{Scope: s}
			}
			return KeyMsg{Key: k}
		case input.Paste:
			return PasteMsg{Text: k.Text}
		default:
			return KeyMsg{Key: k}
		}
	}

	// Message channels
	keyCh := input.ReadKeys(ctx, in)
	resizeCh := input.WatchResize(ctx)
	cmdCh := make(chan Msg, 64)

	// Frame rate limiter
	frameTicker := time.NewTicker(33 * time.Millisecond) // ~30 FPS
	defer frameTicker.Stop()

	needsRender := true
	msgs := make([]Msg, 0, 16)

	for {
		// Collect messages
		select {
		case <-ctx.Done():
			return ctx.Err()
		case k, ok := <-keyCh:
			if !ok {
				return nil
			}
			if m := toMsg(k); m != nil {
				msgs = append(msgs, m)
			}
			needsRender = true
		case r, ok := <-resizeCh:
			if !ok {
				continue
			}
			width, height = r.Width, r.Height
			ansi.ClearScreen(out) // clear stale content in newly exposed areas
			prevBuf = nil         // force full redraw
			msgs = append(msgs, ResizeMsg{Width: width, Height: height})
			needsRender = true
		case cmdMsg := <-cmdCh:
			msgs = append(msgs, cmdMsg)
			needsRender = true
		case <-frameTicker.C:
			// Process batched messages
		}

		// Drain any additional pending messages
		draining := true
		for draining {
			select {
			case k, ok := <-keyCh:
				if !ok {
					draining = false
					continue
				}
				if m := toMsg(k); m != nil {
					msgs = append(msgs, m)
				}
				needsRender = true
			case cmdMsg := <-cmdCh:
				msgs = append(msgs, cmdMsg)
				needsRender = true
			default:
				draining = false
			}
		}

		if !needsRender {
			continue
		}

		// Handle focus keys before update
		for _, msg := range msgs {
			if km, ok := msg.(KeyMsg); ok {
				switch km.Key.Type {
				case input.Tab:
					fm.Next()
				case input.ShiftTab:
					fm.Prev()
				}
			}
		}

		// Process all messages through update
		for _, msg := range msgs {
			result := a.Update(model, msg)
			model = result.Model
			if result.Quit {
				return nil
			}
			// Launch async commands
			for _, cmd := range result.Cmds {
				c := cmd
				go func() {
					if m := c(); m != nil {
						cmdCh <- m
					}
				}()
			}
			// Launch subscriptions
			for _, sub := range result.Subs {
				s := sub
				go func() {
					if m := s(func(msg Msg) { cmdCh <- msg }); m != nil {
						cmdCh <- m
					}
				}()
			}
		}
		msgs = msgs[:0]

		// Render pipeline
		tree := a.View(model, fm.Current())
		lt := layout.Layout(tree, width, height)
		fm.Update(lt)
		lastLayout = &lt

		// Tell Update when focus moved (Tab, click, or scope change) so
		// models can mirror the focused key.
		if cur := fm.Current(); cur != lastFocused {
			lastFocused = cur
			msgs = append(msgs, FocusChangedMsg{Key: cur})
			needsRender = true
		}

		buf := cell.NewBuffer(width, height)
		cell.Paint(buf, lt)

		if prevBuf == nil {
			prevBuf = cell.NewBuffer(width, height) // empty for first frame
		}

		changes := diff.Diff(prevBuf, buf)
		ansi.Render(out, changes)

		prevBuf = buf
		// Keep rendering pending if the frame itself queued messages
		// (e.g. FocusChangedMsg) so they process on the next tick.
		needsRender = len(msgs) > 0
	}
}
