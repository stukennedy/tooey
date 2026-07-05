package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stukennedy/tooey/app"
	"github.com/stukennedy/tooey/input"
	"github.com/stukennedy/tooey/node"

	"golang.org/x/term"
)

type model struct {
	items    []string
	selected int
	counter  int
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	a := &app.App[*model]{
		Init: func() *model {
			return &model{
				items:    []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"},
				selected: 0,
				counter:  0,
			}
		},
		Update: func(mdl *model, msg app.Msg) app.UpdateResult[*model] {
			switch msg := msg.(type) {
			case app.KeyMsg:
				switch msg.Key.Type {
				case input.Up:
					if mdl.selected > 0 {
						mdl.selected--
					}
				case input.Down:
					if mdl.selected < len(mdl.items)-1 {
						mdl.selected++
					}
				case input.Enter:
					mdl.counter++
				case input.RuneKey:
					if msg.Key.Rune == 'q' {
						return app.Quit(mdl)
					}
				}
			}
			return app.NoCmd(mdl)
		},
		View: func(mdl *model, focused string) node.Node {
			items := make([]node.Node, len(mdl.items))
			for i, item := range mdl.items {
				prefix := "  "
				fg := node.Color(7)
				bg := node.Color(0)
				style := node.StyleFlags(0)
				if i == mdl.selected {
					prefix = "> "
					fg = node.Color(0)
					bg = node.Color(6)
					style = node.Bold
				}
				items[i] = node.TextStyled(prefix+item, fg, bg, style)
			}

			title := node.TextStyled(" tooey demo ", node.Color(0), node.Color(2), node.Bold)
			counter := node.Text(fmt.Sprintf(" Activations: %d ", mdl.counter))
			help := node.TextStyled(" ↑/↓ navigate • Enter activate • q quit ", node.Color(8), 0, 0)

			return node.Column(
				title,
				node.Text(""),
				node.Box(node.BorderRounded, node.Column(items...)),
				node.Text(""),
				counter,
				node.Spacer(),
				help,
			)
		},
	}

	if err := a.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
