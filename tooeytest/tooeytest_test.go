package tooeytest

import (
	"testing"

	"github.com/stukennedy/tooey/node"
)

func TestRenderText(t *testing.T) {
	got := RenderText(node.Column(node.Text("hello"), node.Text("world")), 10, 4)
	want := "hello\nworld"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAssertFrameBox(t *testing.T) {
	view := node.Box(node.BorderSingle, node.Text("hi")).WithSize(6, 3)
	AssertFrame(t, view, 10, 4, `
		┌────┐
		│hi  │
		└────┘`)
}

func TestRenderTextWideRunes(t *testing.T) {
	got := RenderText(node.Text("日本"), 10, 1)
	if got != "日本" {
		t.Fatalf("wide runes should appear once each, got %q", got)
	}
}

func TestAssertFrameFailure(t *testing.T) {
	mock := &mockT{}
	AssertFrame(mock, node.Text("actual"), 10, 1, "expected")
	if !mock.failed {
		t.Fatal("AssertFrame should fail on mismatch")
	}
}

type mockT struct {
	testing.TB
	failed bool
}

func (m *mockT) Helper() {}
func (m *mockT) Errorf(format string, args ...any) {
	m.failed = true
}
