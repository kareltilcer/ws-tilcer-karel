// Package ratelimit is a small in-process, per-key fixed-window rate limiter for
// the public write endpoints (contact submit, score submit, link click). karel
// runs as a single instance, so in-memory is sufficient; the key is the client
// IP. A bounded map guards against a spoofed-IP flood (X-Forwarded-For is
// client-influenced).
package ratelimit

import (
	"sync"
	"time"
)

// maxTrackedKeys caps distinct keys held at once; a flood past it wipes the map
// (counters reset) to keep memory bounded.
const maxTrackedKeys = 16384

// Limiter is a fixed-window counter keyed by an arbitrary string.
type Limiter struct {
	mu        sync.Mutex
	max       int
	window    time.Duration
	hits      map[string]*window
	now       func() time.Time
	lastSweep time.Time
}

type window struct {
	start time.Time
	count int
}

// New returns a limiter allowing max events per window per key.
func New(max int, w time.Duration) *Limiter {
	return NewWithClock(max, w, time.Now)
}

// NewWithClock is New with an injectable clock (tests).
func NewWithClock(max int, w time.Duration, now func() time.Time) *Limiter {
	if now == nil {
		now = time.Now
	}
	if max < 1 {
		max = 1
	}
	if w <= 0 {
		w = time.Minute
	}
	return &Limiter{max: max, window: w, hits: map[string]*window{}, now: now}
}

// Allow records an attempt for key and reports whether it is within the limit.
// The attempt is counted whether or not it is allowed (fixed window).
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.sweep(now)
	w := l.hits[key]
	if w == nil || now.Sub(w.start) >= l.window {
		l.hits[key] = &window{start: now, count: 1}
		return true
	}
	w.count++
	return w.count <= l.max
}

// RetryAfter returns how long until key's current window resets (0 if not
// currently limited). Useful for the Retry-After header on a 429.
func (l *Limiter) RetryAfter(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.hits[key]
	if w == nil {
		return 0
	}
	rem := l.window - l.now().Sub(w.start)
	if rem < 0 {
		return 0
	}
	return rem
}

// sweep drops fully-elapsed windows, at most once per window, and wipes the map
// if it exceeds the key cap. Caller holds l.mu.
func (l *Limiter) sweep(now time.Time) {
	overCap := len(l.hits) > maxTrackedKeys
	if !overCap && now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	for k, w := range l.hits {
		if now.Sub(w.start) >= l.window {
			delete(l.hits, k)
		}
	}
	if len(l.hits) > maxTrackedKeys {
		l.hits = map[string]*window{}
	}
}
