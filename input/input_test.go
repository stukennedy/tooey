package input

import "testing"

func TestParseArrowKeys(t *testing.T) {
	tests := []struct {
		input    []byte
		expected KeyType
	}{
		{[]byte{0x1b, '[', 'A'}, Up},
		{[]byte{0x1b, '[', 'B'}, Down},
		{[]byte{0x1b, '[', 'C'}, Right},
		{[]byte{0x1b, '[', 'D'}, Left},
	}
	for _, tt := range tests {
		keys := parseInput(tt.input)
		if len(keys) != 1 || keys[0].Type != tt.expected {
			t.Errorf("input %v: expected %d, got %v", tt.input, tt.expected, keys)
		}
	}
}

func TestParseRunes(t *testing.T) {
	keys := parseInput([]byte("abc"))
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	for i, ch := range "abc" {
		if keys[i].Type != RuneKey || keys[i].Rune != ch {
			t.Errorf("key %d: expected %c, got %v", i, ch, keys[i])
		}
	}
}

func TestParseSpecialKeys(t *testing.T) {
	tests := []struct {
		input    []byte
		expected KeyType
	}{
		{[]byte{'\r'}, Enter},
		{[]byte{'\t'}, Tab},
		{[]byte{0x1b, '[', 'Z'}, ShiftTab},
		{[]byte{0x7f}, Backspace},
		{[]byte{0x03}, CtrlC},
		{[]byte{0x1b}, Escape},
	}
	for _, tt := range tests {
		keys := parseInput(tt.input)
		if len(keys) != 1 || keys[0].Type != tt.expected {
			t.Errorf("input %v: expected %d, got %v", tt.input, tt.expected, keys)
		}
	}
}

func TestParsePageKeys(t *testing.T) {
	keys := parseInput([]byte{0x1b, '[', '5', '~'})
	if len(keys) != 1 || keys[0].Type != PageUp {
		t.Fatalf("expected PageUp, got %v", keys)
	}
}

func TestParseAltArrowKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected KeyType
	}{
		{"AltUp", []byte{0x1b, '[', '1', ';', '3', 'A'}, AltUp},
		{"AltDown", []byte{0x1b, '[', '1', ';', '3', 'B'}, AltDown},
		{"AltRight", []byte{0x1b, '[', '1', ';', '3', 'C'}, AltRight},
		{"AltLeft", []byte{0x1b, '[', '1', ';', '3', 'D'}, AltLeft},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := parseInput(tt.input)
			if len(keys) != 1 || keys[0].Type != tt.expected {
				t.Errorf("input %v: expected %d, got %v", tt.input, tt.expected, keys)
			}
		})
	}
}

func TestUnrecognizedCSIDoesNotEmitEscape(t *testing.T) {
	// \x1b[1;5A = Ctrl+Up — not handled, should be silently skipped
	keys := parseInput([]byte{0x1b, '[', '1', ';', '5', 'A'})
	for _, k := range keys {
		if k.Type == Escape {
			t.Errorf("unrecognized CSI should not emit Escape, got %v", keys)
		}
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys for unrecognized CSI, got %d: %v", len(keys), keys)
	}
}

func TestUnrecognizedCSIFollowedByText(t *testing.T) {
	// Unrecognized CSI then 'x' — should skip the CSI and emit 'x'
	input := append([]byte{0x1b, '[', '1', ';', '5', 'A'}, 'x')
	keys := parseInput(input)
	if len(keys) != 1 || keys[0].Type != RuneKey || keys[0].Rune != 'x' {
		t.Errorf("expected just 'x' after unrecognized CSI, got %v", keys)
	}
}

func TestBracketedPasteComplete(t *testing.T) {
	// \x1b[200~ hello world \x1b[201~
	data := []byte("\x1b[200~hello world\x1b[201~")
	keys := parseInput(data)
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(keys), keys)
	}
	if keys[0].Type != Paste {
		t.Errorf("expected Paste key type, got %d", keys[0].Type)
	}
	if keys[0].Text != "hello world" {
		t.Errorf("expected 'hello world', got %q", keys[0].Text)
	}
}

func TestBracketedPasteWithSurroundingKeys(t *testing.T) {
	// 'a' then paste then 'b'
	data := []byte("a\x1b[200~pasted\x1b[201~b")
	keys := parseInput(data)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}
	if keys[0].Type != RuneKey || keys[0].Rune != 'a' {
		t.Errorf("key 0: expected 'a', got %v", keys[0])
	}
	if keys[1].Type != Paste || keys[1].Text != "pasted" {
		t.Errorf("key 1: expected Paste 'pasted', got %v", keys[1])
	}
	if keys[2].Type != RuneKey || keys[2].Rune != 'b' {
		t.Errorf("key 2: expected 'b', got %v", keys[2])
	}
}

func TestBracketedPasteMultiline(t *testing.T) {
	// Paste with newlines
	data := []byte("\x1b[200~line1\nline2\nline3\x1b[201~")
	keys := parseInput(data)
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(keys), keys)
	}
	if keys[0].Type != Paste {
		t.Errorf("expected Paste key type, got %d", keys[0].Type)
	}
	if keys[0].Text != "line1\nline2\nline3" {
		t.Errorf("expected multiline text, got %q", keys[0].Text)
	}
}

func TestBracketedPastePartial(t *testing.T) {
	// Paste start with no end marker — returns no keys (partial paste)
	data := []byte("\x1b[200~partial content")
	keys := parseInput(data)
	// parseInput returns no keys for partial paste (ReadKeys handles buffering)
	if len(keys) != 0 {
		t.Errorf("expected 0 keys for partial paste, got %d: %v", len(keys), keys)
	}
}

func TestBracketedPastePartialWithPrecedingKeys(t *testing.T) {
	// Keys before a partial paste
	data := []byte("abc\x1b[200~partial")
	keys := parseInput(data)
	// Should get 3 rune keys for 'a', 'b', 'c', then stop at the partial paste
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys before partial paste, got %d: %v", len(keys), keys)
	}
	for i, ch := range "abc" {
		if keys[i].Type != RuneKey || keys[i].Rune != ch {
			t.Errorf("key %d: expected %c, got %v", i, ch, keys[i])
		}
	}
}

func TestBracketedPasteEmpty(t *testing.T) {
	// Empty paste
	data := []byte("\x1b[200~\x1b[201~")
	keys := parseInput(data)
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d: %v", len(keys), keys)
	}
	if keys[0].Type != Paste || keys[0].Text != "" {
		t.Errorf("expected empty Paste, got %v", keys[0])
	}
}
