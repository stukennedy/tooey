package input

import "testing"

func TestParseSGRMouseClick(t *testing.T) {
	// Left button press at column 10, row 20 (1-based)
	keys := parseInput([]byte("\x1b[<0;10;20M"))
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	k := keys[0]
	if k.Type != MouseClick {
		t.Fatalf("expected MouseClick, got %v", k.Type)
	}
	if k.MouseX != 9 || k.MouseY != 19 {
		t.Fatalf("expected 0-based (9,19), got (%d,%d)", k.MouseX, k.MouseY)
	}
}

func TestParseSGRMouseRelease(t *testing.T) {
	keys := parseInput([]byte("\x1b[<0;5;6m"))
	if len(keys) != 1 || keys[0].Type != MouseRelease {
		t.Fatalf("expected MouseRelease, got %+v", keys)
	}
}

func TestParseSGRMouseScroll(t *testing.T) {
	keys := parseInput([]byte("\x1b[<64;3;4M\x1b[<65;3;4M"))
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Type != MouseScrollUp || keys[1].Type != MouseScrollDown {
		t.Fatalf("expected scroll up+down, got %+v", keys)
	}
	if keys[0].MouseX != 2 || keys[0].MouseY != 3 {
		t.Fatalf("scroll should carry coordinates, got (%d,%d)", keys[0].MouseX, keys[0].MouseY)
	}
}

func TestParseNormalMouseClick(t *testing.T) {
	// X10 encoding: btn+32, x+33, y+33 → click at 0-based (7, 11)
	keys := parseInput([]byte{0x1b, '[', 'M', 32, 40, 44})
	if len(keys) != 1 || keys[0].Type != MouseClick {
		t.Fatalf("expected MouseClick, got %+v", keys)
	}
	if keys[0].MouseX != 7 || keys[0].MouseY != 11 {
		t.Fatalf("expected (7,11), got (%d,%d)", keys[0].MouseX, keys[0].MouseY)
	}
}
