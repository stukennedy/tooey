package app

import "testing"

func TestWithSub(t *testing.T) {
	sub := func(send func(Msg)) Msg {
		send("intermediate")
		return "done"
	}
	result := WithSub("model", sub)
	if result.Model != "model" {
		t.Fatal("unexpected model")
	}
	if len(result.Subs) != 1 {
		t.Fatalf("expected 1 sub, got %d", len(result.Subs))
	}

	// Verify the sub works correctly
	var received []Msg
	final := result.Subs[0](func(msg Msg) {
		received = append(received, msg)
	})
	if len(received) != 1 || received[0] != "intermediate" {
		t.Fatalf("expected 1 intermediate msg, got %v", received)
	}
	if final != "done" {
		t.Fatalf("expected final 'done', got %v", final)
	}
}

func TestNoCmd(t *testing.T) {
	r := NoCmd("m")
	if r.Model != "m" || len(r.Cmds) != 0 || len(r.Subs) != 0 {
		t.Fatal("unexpected result")
	}
}

func TestWithCmd(t *testing.T) {
	cmd := func() Msg { return "msg" }
	r := WithCmd("m", cmd)
	if len(r.Cmds) != 1 {
		t.Fatalf("expected 1 cmd, got %d", len(r.Cmds))
	}
}
