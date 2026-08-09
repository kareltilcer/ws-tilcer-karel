package auth

import (
	"sync"
	"time"
)

// maxTrackedKeys caps distinct login rate-limit keys; a spoofed-key flood wipes
// the map (counters reset) to keep memory bounded.
const maxTrackedKeys = 4096

// loginLimiter is a fixed-window counter keyed by an arbitrary string. Login uses
// one per IP and one per (IP,email). Only FAILED logins are counted (via fail),
// so a user is never locked out by their own successful sign-ins.
type loginLimiter struct {
	mu        sync.Mutex
	max       int
	window    time.Duration
	hits      map[string]*hitWindow
	now       func() time.Time
	lastSweep time.Time
}

type hitWindow struct {
	start time.Time
	count int
}

func newLoginLimiter(max int, window time.Duration, now func() time.Time) *loginLimiter {
	if now == nil {
		now = time.Now
	}
	return &loginLimiter{max: max, window: window, hits: map[string]*hitWindow{}, now: now}
}

// allowed reports whether key is currently under the limit, WITHOUT recording.
func (r *loginLimiter) allowed(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	w := r.hits[key]
	if w == nil || r.now().Sub(w.start) >= r.window {
		return true
	}
	return w.count < r.max
}

// fail records a failed attempt for key (fixed-window).
func (r *loginLimiter) fail(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	r.sweep(now)
	w := r.hits[key]
	if w == nil || now.Sub(w.start) >= r.window {
		r.hits[key] = &hitWindow{start: now, count: 1}
		return
	}
	w.count++
}

func (r *loginLimiter) sweep(now time.Time) {
	overCap := len(r.hits) > maxTrackedKeys
	if !overCap && now.Sub(r.lastSweep) < r.window {
		return
	}
	r.lastSweep = now
	for k, w := range r.hits {
		if now.Sub(w.start) >= r.window {
			delete(r.hits, k)
		}
	}
	if len(r.hits) > maxTrackedKeys {
		r.hits = map[string]*hitWindow{}
	}
}

// reset clears the counter for key after a successful login.
func (r *loginLimiter) reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hits, key)
}
