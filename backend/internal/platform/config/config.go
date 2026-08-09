// Package config loads and validates the service configuration from the
// environment (PRD §9). It fails fast and loudly: a missing required variable or
// an invalid value aborts startup with a message listing every problem.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Rate is a parsed "N/unit" rate limit (e.g. "5/min" → Max=5, Window=1m).
type Rate struct {
	Max    int
	Window time.Duration
}

// Config is the fully-validated runtime configuration.
type Config struct {
	// --- core service ---
	Env       string // "development" (default) | "production"; gates the dev bypass
	Addr      string // TCP listen address, e.g. ":2002"
	DBPath    string // SQLite file path (persisted volume in production)
	StaticDir string // built SPA dir; empty in the two-app deploy (SPA is its own image)
	SiteKey   string // auth site key ("karel")

	// --- Mode B auth (mirror home) ---
	AuthBaseURL        string   // shared auth service base URL (e.g. https://auth.tilcer.cz/api)
	AuthServiceSecret  string   // X-Service-Secret (BE→BE); never logged
	AuthJWTSecret      string   // shared HS256 secret to VERIFY returned access tokens; never logged
	AuthJWTIssuer      string   // expected `iss` (auth's own base URL); "" = do not enforce
	AllowedOrigins     []string // CSRF Origin allowlist for admin mutations
	SessionTTLDays     int      // sliding session window (default 90)
	RoleRefreshMinutes int      // role re-mint threshold (default 15)

	// --- dev bypass (never in production) ---
	DevAuthBypass bool
	DevActorID    string
	DevActorRoles []string

	// --- contact email (Resend) ---
	ResendAPIKey string
	ContactTo    string
	ContactFrom  string

	// --- spam protection ---
	TurnstileSecret string // "" disables Turnstile verification (dev); honeypot still enforced

	// --- anti-abuse / caps ---
	ContactRate     Rate
	ScoreRate       Rate
	ClickRate       Rate
	MaxContactBytes int64
	MaxMediaBytes   int64

	// --- media S3 (self-hosted, Coolify) ---
	S3Endpoint         string
	S3Region           string
	S3Bucket           string
	S3AccessKey        string
	S3SecretKey        string
	S3ForcePathStyle   bool
	S3MediaPrefix      string // object key prefix (default "media/")
	MediaPublicBaseURL string // public base for served media URLs
}

// IsProduction reports whether the service is running in production.
func (c *Config) IsProduction() bool { return c.Env == "production" }

// Redacted returns a log-safe one-line summary with secrets masked.
func (c *Config) Redacted() string {
	mask := func(s string) string {
		if s == "" {
			return "unset"
		}
		return "set(***)"
	}
	static := c.StaticDir
	if static == "" {
		static = "none"
	}
	return fmt.Sprintf(
		"env=%s addr=%s db=%s static=%s site=%s auth_base=%s auth_secret=%s jwt_secret=%s "+
			"origins=%v session_ttl_days=%d role_refresh_min=%d dev_bypass=%t "+
			"resend=%s turnstile=%s s3_bucket=%s media_base=%s",
		c.Env, c.Addr, c.DBPath, static, c.SiteKey, c.AuthBaseURL, mask(c.AuthServiceSecret),
		mask(c.AuthJWTSecret), c.AllowedOrigins, c.SessionTTLDays, c.RoleRefreshMinutes, c.DevAuthBypass,
		mask(c.ResendAPIKey), mask(c.TurnstileSecret), c.S3Bucket, c.MediaPublicBaseURL,
	)
}

// Getenv mirrors os.LookupEnv and is injected in tests.
type Getenv func(key string) (string, bool)

const (
	defaultEnv                = "development"
	defaultAddr               = ":2002"
	defaultSiteKey            = "karel"
	defaultSessionTTLDays     = 90
	defaultRoleRefreshMinutes = 15
	defaultMaxContactBytes    = 32 * 1024        // 32 KiB
	defaultMaxMediaBytes      = 10 * 1024 * 1024 // 10 MiB
	defaultS3MediaPrefix      = "media/"
)

var defaultAllowedOrigins = []string{"https://*.tilcer.cz"}

// LoadFromEnv loads the configuration using os.LookupEnv.
func LoadFromEnv() (*Config, error) { return Load(osLookup) }

// Load reads, defaults, and validates configuration from getenv. On any problem
// it returns a single error enumerating every issue found.
func Load(getenv Getenv) (*Config, error) {
	l := &loader{getenv: getenv}
	c := &Config{}

	c.Env = l.strDefault("KAREL_ENV", defaultEnv)
	if c.Env != "development" && c.Env != "production" {
		l.errf("KAREL_ENV must be \"development\" or \"production\" (got %q)", c.Env)
	}
	c.Addr = l.strDefault("KAREL_ADDR", defaultAddr)
	c.DBPath = l.strRequired("KAREL_DB_PATH")
	c.StaticDir = l.strDefault("KAREL_STATIC_DIR", "")
	c.SiteKey = l.strDefault("KAREL_SITE_KEY", defaultSiteKey)

	c.DevAuthBypass = l.boolDefault("KAREL_DEV_AUTH_BYPASS", false)
	c.DevActorID = l.strDefault("KAREL_DEV_ACTOR_ID", "dev-admin")
	c.DevActorRoles = l.csvDefault("KAREL_DEV_ACTOR_ROLES", []string{"admin"})

	// The auth service is only strictly required when the bypass is off.
	if c.DevAuthBypass {
		c.AuthBaseURL = l.strDefault("AUTH_BASE_URL", "")
		c.AuthServiceSecret = l.strDefault("KAREL_AUTH_SERVICE_SECRET", "")
		c.AuthJWTSecret = l.strDefault("KAREL_AUTH_JWT_SECRET", "")
	} else {
		c.AuthBaseURL = l.strRequired("AUTH_BASE_URL")
		c.AuthServiceSecret = l.strRequired("KAREL_AUTH_SERVICE_SECRET")
		c.AuthJWTSecret = l.strRequired("KAREL_AUTH_JWT_SECRET")
	}
	c.AuthJWTIssuer = l.strDefault("KAREL_AUTH_JWT_ISSUER", "")
	c.AllowedOrigins = l.csvDefault("KAREL_ALLOWED_ORIGINS", defaultAllowedOrigins)
	c.SessionTTLDays = l.intDefault("KAREL_SESSION_TTL_DAYS", defaultSessionTTLDays)
	c.RoleRefreshMinutes = l.intDefault("KAREL_ROLE_REFRESH_MINUTES", defaultRoleRefreshMinutes)

	// Contact email — optional at boot (validated when contact actually sends).
	c.ResendAPIKey = l.strDefault("RESEND_API_KEY", "")
	c.ContactTo = l.strDefault("CONTACT_TO", "")
	c.ContactFrom = l.strDefault("CONTACT_FROM", "")

	c.TurnstileSecret = l.strDefault("TURNSTILE_SECRET", "")

	c.ContactRate = l.rateDefault("CONTACT_RATE", Rate{Max: 5, Window: time.Minute})
	c.ScoreRate = l.rateDefault("SCORE_RATE", Rate{Max: 20, Window: time.Minute})
	c.ClickRate = l.rateDefault("CLICK_RATE", Rate{Max: 60, Window: time.Minute})
	c.MaxContactBytes = int64(l.intDefault("MAX_CONTACT_BYTES", defaultMaxContactBytes))
	c.MaxMediaBytes = int64(l.intDefault("MAX_MEDIA_BYTES", defaultMaxMediaBytes))

	// Media S3 — optional at boot (validated when media actually uploads).
	c.S3Endpoint = l.strDefault("S3_ENDPOINT", "")
	c.S3Region = l.strDefault("S3_REGION", "auto")
	c.S3Bucket = l.strDefault("S3_BUCKET", "")
	c.S3AccessKey = l.strDefault("S3_ACCESS_KEY", "")
	c.S3SecretKey = l.strDefault("S3_SECRET_KEY", "")
	c.S3ForcePathStyle = l.boolDefault("S3_FORCE_PATH_STYLE", true)
	c.S3MediaPrefix = l.strDefault("S3_MEDIA_PREFIX", defaultS3MediaPrefix)
	c.MediaPublicBaseURL = l.strDefault("MEDIA_PUBLIC_BASE_URL", "")

	// Range sanity.
	if c.SessionTTLDays < 1 {
		l.errf("KAREL_SESSION_TTL_DAYS must be >= 1 (got %d)", c.SessionTTLDays)
	}
	if c.RoleRefreshMinutes < 1 {
		l.errf("KAREL_ROLE_REFRESH_MINUTES must be >= 1 (got %d)", c.RoleRefreshMinutes)
	}
	if c.MaxContactBytes < 1 {
		l.errf("MAX_CONTACT_BYTES must be >= 1 (got %d)", c.MaxContactBytes)
	}
	if c.MaxMediaBytes < 1 {
		l.errf("MAX_MEDIA_BYTES must be >= 1 (got %d)", c.MaxMediaBytes)
	}

	// Security hard-stop: the dev bypass must never be active in production.
	if c.DevAuthBypass && c.IsProduction() {
		l.errf("KAREL_DEV_AUTH_BYPASS must not be enabled when KAREL_ENV=production " +
			"(fake authentication in production is a security hole)")
	}

	if len(l.errs) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(l.errs, "\n  - "))
	}
	return c, nil
}

// MediaConfigured reports whether the S3 media store is fully configured.
func (c *Config) MediaConfigured() bool {
	return c.S3Endpoint != "" && c.S3Bucket != "" && c.S3AccessKey != "" &&
		c.S3SecretKey != "" && c.MediaPublicBaseURL != ""
}

// EmailConfigured reports whether Resend is fully configured.
func (c *Config) EmailConfigured() bool {
	return c.ResendAPIKey != "" && c.ContactTo != "" && c.ContactFrom != ""
}

func osLookup(key string) (string, bool) { return os.LookupEnv(key) }

// loader accumulates validation errors while reading typed values.
type loader struct {
	getenv Getenv
	errs   []string
}

func (l *loader) errf(format string, a ...any) { l.errs = append(l.errs, fmt.Sprintf(format, a...)) }

func (l *loader) strRequired(key string) string {
	v, ok := l.getenv(key)
	if !ok || strings.TrimSpace(v) == "" {
		l.errf("%s is required", key)
		return ""
	}
	return v
}

func (l *loader) strDefault(key, def string) string {
	v, ok := l.getenv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func (l *loader) intDefault(key string, def int) int {
	v, ok := l.getenv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		l.errf("%s must be an integer (got %q)", key, v)
		return def
	}
	return n
}

func (l *loader) boolDefault(key string, def bool) bool {
	v, ok := l.getenv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		l.errf("%s must be a boolean (got %q)", key, v)
		return def
	}
	return b
}

func (l *loader) csvDefault(key string, def []string) []string {
	v, ok := l.getenv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

// rateDefault parses a "N/unit" rate (e.g. "5/min", "10/s", "100/hour"). unit ∈
// {s|sec|second, m|min|minute, h|hour}. Returns def when unset or empty.
func (l *loader) rateDefault(key string, def Rate) Rate {
	v, ok := l.getenv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	n, w, err := parseRate(strings.TrimSpace(v))
	if err != nil {
		l.errf("%s must be like \"5/min\" (got %q): %v", key, v, err)
		return def
	}
	return Rate{Max: n, Window: w}
}

func parseRate(s string) (int, time.Duration, error) {
	num, unit, found := strings.Cut(s, "/")
	if !found {
		return 0, 0, fmt.Errorf("missing '/'")
	}
	n, err := strconv.Atoi(strings.TrimSpace(num))
	if err != nil || n < 1 {
		return 0, 0, fmt.Errorf("count must be a positive integer")
	}
	var w time.Duration
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "s", "sec", "second":
		w = time.Second
	case "m", "min", "minute":
		w = time.Minute
	case "h", "hour":
		w = time.Hour
	default:
		return 0, 0, fmt.Errorf("unknown unit %q", unit)
	}
	return n, w, nil
}
