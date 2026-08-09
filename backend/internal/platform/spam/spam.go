// Package spam holds the shared anti-abuse checks for the public write endpoints
// (contact submit, score submit): a honeypot field that must stay empty and a
// Cloudflare Turnstile token verified server-side. Both /api/contact and
// /api/arcade/{gameKey}/scores run these before persisting anything.
package spam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Guard performs the honeypot + Turnstile checks.
type Guard struct {
	secret string // Turnstile secret; "" disables verification (dev/local)
	client *http.Client
	// verifyURL is Cloudflare's siteverify endpoint (overridable in tests).
	verifyURL string
}

// NewGuard returns a Guard. When secret is empty, Turnstile verification is
// skipped (honeypot is still enforced) — for local development only.
func NewGuard(secret string) *Guard {
	return &Guard{
		secret:    secret,
		client:    &http.Client{Timeout: 5 * time.Second},
		verifyURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify",
	}
}

// HoneypotFilled reports whether the honeypot field carries a value (any
// non-empty value means a bot filled a field a human never sees).
func HoneypotFilled(value string) bool { return strings.TrimSpace(value) != "" }

// Enabled reports whether Turnstile verification is active.
func (g *Guard) Enabled() bool { return g.secret != "" }

// Check runs the full guard: honeypot must be empty and, when Turnstile is
// enabled, the token must verify against Cloudflare. Returns nil when the
// request passes, or a non-nil error describing the failure (callers map any
// failure to a 400 spam_rejected). remoteIP is optional (improves scoring).
func (g *Guard) Check(ctx context.Context, honeypot, token, remoteIP string) error {
	if HoneypotFilled(honeypot) {
		return errHoneypot
	}
	if !g.Enabled() {
		return nil // dev/local: skip network verification
	}
	if strings.TrimSpace(token) == "" {
		return errNoToken
	}
	ok, err := g.verify(ctx, token, remoteIP)
	if err != nil {
		return err
	}
	if !ok {
		return errTurnstile
	}
	return nil
}

type siteverifyResponse struct {
	Success bool `json:"success"`
}

func (g *Guard) verify(ctx context.Context, token, remoteIP string) (bool, error) {
	form := url.Values{}
	form.Set("secret", g.secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := g.client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()
	var out siteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Success, nil
}

// sentinel errors (callers only need to know it failed → 400).
type spamError string

func (e spamError) Error() string { return string(e) }

const (
	errHoneypot  = spamError("honeypot field filled")
	errNoToken   = spamError("missing turnstile token")
	errTurnstile = spamError("turnstile verification failed")
)
