# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Tooey — React-style, server-driven terminal UI runtime in Go. Declarative component trees rendered to a cell buffer with efficient frame diffing.

## Commands

```bash
go test ./...                  # Run all tests
go test ./layout               # Run tests for a single package
go test ./cell -run TestPaint  # Run a single test
go run ./demos/list            # Interactive list demo
go run ./demos/maude           # Claude Code-style chat TUI demo
```

## Architecture

```
Input → Update Loop → Render Tree → Layout → Cell Buffer → Diff → ANSI Output
```

Elm-style: UI is a pure function of state. No imperative widget mutation. The app loop (`app/app.go`) collects input messages, batches them per frame (~30fps), runs Update for each, then executes the render pipeline: `View()` → `layout.Layout()` → `cell.Paint()` → `diff.Diff()` → `ansi.Render()`.

### Render pipeline detail

1. `View` returns a `node.Node` tree (value structs, not interfaces)
2. `layout.Layout()` produces a `LayoutNode` tree with computed positions/sizes (single-pass flex layout)
3. `cell.Paint()` renders the layout tree into a flat `[]Cell` buffer (row-major)
4. `diff.Diff()` compares previous and current buffers, grouping adjacent changes into minimal runs
5. `ansi.Render()` emits ANSI escape sequences for the diff

### Packages

| Package | Purpose |
|---------|---------|
| `node` | Node tree types (Text, Box, Row, Column, List, Pane, Spacer, Overlay), builder funcs, Color model (default/ANSI-256/RGB) |
| `textwidth` | Display-width measurement (wcwidth-style): CJK/emoji = 2 cells, combining marks = 0 |
| `cell` | Cell buffer (`[]Cell` with Rune/FG/BG/Style), wide-rune invariants, `Paint()` renders layout to buffer |
| `layout` | Single-pass flex layout engine (measure + place), padding, scroll offset, `HitTest` for mouse |
| `diff` | Cell-by-cell frame diff, groups adjacent changes into minimal runs |
| `ansi` | ANSI escape sequence emitter (256/truecolor with downgrade), alt screen, cursor control |
| `input` | Raw terminal key + mouse parsing (SGR coordinates), resize detection via SIGWINCH |
| `focus` | Focus manager with Tab cycling, direct `Focus(key)`, push/pop context stack |
| `app` | Elm-style main loop, generic `App[M]` (Init/Update/View), async Cmd system, click hit-testing, 30fps |
| `sse` | SSE client + HTTP POST for server integration |
| `wire` | JSON serialization of node trees + actions for server-driven UIs |
| `component` | Reusable components: TextInput, List, Table, Tabs, Select, Progress, Badge, Spinner, Steps, Collapsible |
| `tooeytest` | Golden-frame test helpers (`RenderText`, `AssertFrame`) |

## Key Decisions

- Node is a value struct, not an interface — builder funcs with chaining (`.WithKey()`, `.WithFlex()`, `.WithPadding()`, etc.)
- Buffer is flat `[]Cell` (row-major) — cache-friendly, simple to diff; wide runes own a continuation cell (Rune 0)
- Color is a uint32 scalar: 0 = default, 1–255 = palette, `Ansi(n)`/`RGB(r,g,b)` for explicit palette/truecolor
- Layout produces a separate `LayoutNode` tree — keeps virtual tree immutable; also used for mouse hit-testing
- Focus managed outside component tree — rebuilt each frame from layout tree; clicks focus via `HitTest`
- Overlay = children stacked in the same rect, painted in order (no z-buffer needed)
- SSE decoupled from render — just another Msg source, batched per frame
- App/Update are generic over the model: `app.App[M]`, `app.NoCmd(model)`, `app.WithCmd(model, cmds...)`, `app.Quit(model)` to quit
- Only external dependency: `golang.org/x/term`
