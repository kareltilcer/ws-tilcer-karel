package config

import (
	"testing"
	"time"
)

func envFrom(m map[string]string) Getenv {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoadDefaults(t *testing.T) {
	c, err := Load(envFrom(map[string]string{
		"KAREL_DB_PATH":         "/data/karel.db",
		"KAREL_DEV_AUTH_BYPASS": "true",
	}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Addr != ":2002" {
		t.Errorf("Addr = %q, want :2002", c.Addr)
	}
	if c.SiteKey != "karel" {
		t.Errorf("SiteKey = %q, want karel", c.SiteKey)
	}
	if c.ContactRate.Max != 5 || c.ContactRate.Window != time.Minute {
		t.Errorf("ContactRate = %+v, want 5/min", c.ContactRate)
	}
	if c.MediaConfigured() {
		t.Error("MediaConfigured() = true with no S3 env")
	}
	if c.EmailConfigured() {
		t.Error("EmailConfigured() = true with no Resend env")
	}
}

func TestLoadRequiresAuthWithoutBypass(t *testing.T) {
	_, err := Load(envFrom(map[string]string{
		"KAREL_DB_PATH": "/data/karel.db",
		// no AUTH_* and bypass off → must error
	}))
	if err == nil {
		t.Fatal("expected error for missing auth config without bypass")
	}
}

func TestBypassInProductionRejected(t *testing.T) {
	_, err := Load(envFrom(map[string]string{
		"KAREL_DB_PATH":         "/data/karel.db",
		"KAREL_ENV":             "production",
		"KAREL_DEV_AUTH_BYPASS": "true",
	}))
	if err == nil {
		t.Fatal("expected error: dev bypass must be rejected in production")
	}
}

func TestParseRate(t *testing.T) {
	cases := map[string]struct {
		n  int
		w  time.Duration
		ok bool
	}{
		"5/min":    {5, time.Minute, true},
		"10/s":     {10, time.Second, true},
		"100/hour": {100, time.Hour, true},
		"3/minute": {3, time.Minute, true},
		"bad":      {0, 0, false},
		"0/min":    {0, 0, false},
		"5/week":   {0, 0, false},
	}
	for in, want := range cases {
		n, w, err := parseRate(in)
		if want.ok && err != nil {
			t.Errorf("parseRate(%q) unexpected error: %v", in, err)
			continue
		}
		if !want.ok {
			if err == nil {
				t.Errorf("parseRate(%q) expected error, got %d/%v", in, n, w)
			}
			continue
		}
		if n != want.n || w != want.w {
			t.Errorf("parseRate(%q) = %d/%v, want %d/%v", in, n, w, want.n, want.w)
		}
	}
}
