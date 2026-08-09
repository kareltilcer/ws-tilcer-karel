package ratelimit

import (
	"testing"
	"time"
)

func TestAllowWithinWindow(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewWithClock(3, time.Minute, func() time.Time { return now })

	for i := 1; i <= 3; i++ {
		if !l.Allow("ip") {
			t.Fatalf("attempt %d should be allowed", i)
		}
	}
	if l.Allow("ip") {
		t.Fatal("4th attempt within window should be blocked")
	}

	// A different key is independent.
	if !l.Allow("other") {
		t.Fatal("independent key should be allowed")
	}

	// After the window elapses the counter resets.
	now = now.Add(time.Minute + time.Second)
	if !l.Allow("ip") {
		t.Fatal("attempt after window reset should be allowed")
	}
}

func TestRetryAfter(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewWithClock(1, time.Minute, func() time.Time { return now })
	l.Allow("ip")
	ra := l.RetryAfter("ip")
	if ra <= 0 || ra > time.Minute {
		t.Fatalf("RetryAfter = %v, want (0, 1m]", ra)
	}
	if l.RetryAfter("unknown") != 0 {
		t.Fatal("RetryAfter for unknown key should be 0")
	}
}
