package spam

import (
	"context"
	"testing"
)

func TestHoneypotFilled(t *testing.T) {
	if !HoneypotFilled("bot") {
		t.Error("non-empty honeypot should be detected")
	}
	if HoneypotFilled("") || HoneypotFilled("   ") {
		t.Error("empty/whitespace honeypot should pass")
	}
}

func TestCheckHoneypotRejected(t *testing.T) {
	g := NewGuard("") // Turnstile disabled
	if err := g.Check(context.Background(), "i-am-a-bot", "", ""); err == nil {
		t.Fatal("filled honeypot must be rejected even with Turnstile disabled")
	}
}

func TestCheckDisabledTurnstilePasses(t *testing.T) {
	g := NewGuard("") // Turnstile disabled → honeypot-only
	if g.Enabled() {
		t.Fatal("guard should report Turnstile disabled with empty secret")
	}
	if err := g.Check(context.Background(), "", "", ""); err != nil {
		t.Fatalf("empty honeypot with Turnstile disabled should pass: %v", err)
	}
}
