// Package focus tracks which node is focused. Focus state is rebuilt
// from the layout tree every frame, so it is a pure function of the
// rendered view plus the user's Tab/click history.
//
// A node with Props.FocusScope (see node.WithFocusScope) traps focus:
// while the scope is rendered, only focusables inside the topmost scope
// participate in cycling. Opening a scope saves the current focus and
// removing it restores it, so modals get correct focus behavior just by
// appearing in and disappearing from the view.
package focus

import (
	"strconv"

	"github.com/stukennedy/tooey/layout"
)

// Manager tracks focus state across the component tree.
type Manager struct {
	focusables []string // ordered keys of focusable nodes in the active scope
	current    int
	scopeKey   string       // identity of the active scope ("" = whole tree)
	saved      []savedFocus // focus to restore when scopes close, innermost last
}

type savedFocus struct {
	scopeKey string // the scope that was active when this was saved
	focusKey string // the focused key to restore on return to it
}

// NewManager creates a new focus manager.
func NewManager() *Manager {
	return &Manager{}
}

// Update rebuilds focus state from the layout tree. Preserves the
// current focus if its key still exists in the active scope. When the
// active focus scope changes, focus is saved (scope opened) or
// restored (scope closed).
func (m *Manager) Update(tree layout.LayoutNode) {
	scopeRoot, scopeKey := activeScope(&tree)
	oldKey := m.Current()

	if scopeKey != m.scopeKey {
		if idx := m.findSaved(scopeKey); idx >= 0 {
			// Returning to a scope we were in before: restore its focus.
			oldKey = m.saved[idx].focusKey
			m.saved = m.saved[:idx]
		} else {
			// Entering a new scope: remember where we were.
			m.saved = append(m.saved, savedFocus{scopeKey: m.scopeKey, focusKey: oldKey})
			oldKey = "" // start at the scope's first focusable
		}
		m.scopeKey = scopeKey
	}

	m.focusables = m.focusables[:0]
	collectFocusables(*scopeRoot, &m.focusables)

	m.current = 0
	if oldKey != "" {
		for i, k := range m.focusables {
			if k == oldKey {
				m.current = i
				break
			}
		}
	}
}

// findSaved returns the index of the most recent saved entry for the
// given scope, or -1.
func (m *Manager) findSaved(scopeKey string) int {
	for i := len(m.saved) - 1; i >= 0; i-- {
		if m.saved[i].scopeKey == scopeKey {
			return i
		}
	}
	return -1
}

// activeScope returns the subtree that owns focus — the topmost focus
// scope (last in paint order), or the whole tree if none — along with
// the scope's identity key.
func activeScope(tree *layout.LayoutNode) (*layout.LayoutNode, string) {
	root, key := tree, ""
	var walk func(ln *layout.LayoutNode, path string)
	walk = func(ln *layout.LayoutNode, path string) {
		if ln.Node.Props.FocusScope {
			root = ln
			key = ln.Node.Props.Key
			if key == "" {
				// Fall back to the tree position for identity; give
				// scopes a Key to keep nesting stable across re-renders.
				key = "scope@" + path
			}
		}
		for i := range ln.Children {
			walk(&ln.Children[i], path+"."+strconv.Itoa(i))
		}
	}
	walk(tree, "0")
	return root, key
}

// Current returns the key of the currently focused node, or "" if none.
func (m *Manager) Current() string {
	if len(m.focusables) == 0 {
		return ""
	}
	if m.current >= len(m.focusables) {
		m.current = 0
	}
	return m.focusables[m.current]
}

// Focus moves focus directly to the node with the given key, returning
// true if the key is focusable in the active scope.
func (m *Manager) Focus(key string) bool {
	for i, k := range m.focusables {
		if k == key {
			m.current = i
			return true
		}
	}
	return false
}

// Next moves focus to the next focusable node (Tab).
func (m *Manager) Next() {
	if len(m.focusables) == 0 {
		return
	}
	m.current = (m.current + 1) % len(m.focusables)
}

// Prev moves focus to the previous focusable node (Shift+Tab).
func (m *Manager) Prev() {
	if len(m.focusables) == 0 {
		return
	}
	m.current = (m.current - 1 + len(m.focusables)) % len(m.focusables)
}

// FocusableCount returns the number of focusable nodes in the active scope.
func (m *Manager) FocusableCount() int {
	return len(m.focusables)
}

func collectFocusables(ln layout.LayoutNode, keys *[]string) {
	if ln.Node.Props.Focusable && ln.Node.Props.Key != "" {
		*keys = append(*keys, ln.Node.Props.Key)
	}
	for _, child := range ln.Children {
		collectFocusables(child, keys)
	}
}
