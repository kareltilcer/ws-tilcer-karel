package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// httpAuthenticator talks to the shared auth service BE→BE as the `karel` service
// client, authenticating with the X-Service-Secret header. Wire contract
// (confirmed against ws-tilcer-auth + ws-tilcer-home): both endpoints return a
// TokenResponse — a site-scoped HS256 access token, NOT a plain identity
// envelope. The user id, email, display name, and roles live in the JWT CLAIMS
// (sub / email / name / roles), so karel verifies the token and reads the claims.
//
//	POST {base}/internal/login
//	  headers: X-Service-Secret: <secret>
//	  body:    {"email": "...", "password": "...", "site": "karel"}
//	  200 -> {"access_token": "<JWT>", "token_type": "Bearer", "expires_in": n, "site": "karel"}
//	         (or {"mfa_required": true, ...}; karel does not handle MFA)
//	  401 -> bad credentials · 403 -> disabled/no access · 5xx -> unreachable
//
//	POST {base}/internal/token/mint
//	  headers: X-Service-Secret: <secret>
//	  body:    {"user_id": "...", "site": "karel"}
//	  200 -> {"access_token": "<JWT>", ...}
//	  403/404/409 -> user closed · 5xx -> unreachable (transient)
type httpAuthenticator struct {
	baseURL   string
	secret    string
	jwtSecret []byte
	issuer    string
	site      string
	client    *http.Client
	logger    *slog.Logger
}

// NewHTTPAuthenticator returns an Authenticator backed by the auth service.
// jwtSecret must equal the auth service's HS256 signing secret. issuer, when
// non-empty, is the exact `iss` claim required (auth stamps its OWN base URL,
// which differs from baseURL); empty disables the issuer check.
func NewHTTPAuthenticator(baseURL, serviceSecret, jwtSecret, issuer, site string, logger *slog.Logger) Authenticator {
	return &httpAuthenticator{
		baseURL:   baseURL,
		secret:    serviceSecret,
		jwtSecret: []byte(jwtSecret),
		issuer:    issuer,
		site:      site,
		client:    &http.Client{Timeout: 5 * time.Second},
		logger:    logger,
	}
}

func (h *httpAuthenticator) log() *slog.Logger {
	if h.logger != nil {
		return h.logger
	}
	return slog.Default()
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Site        string `json:"site"`
	MFARequired bool   `json:"mfa_required"`
}

type accessClaims struct {
	Site  string   `json:"site"`
	Roles []string `json:"roles"`
	Email string   `json:"email"`
	Name  string   `json:"name"`
	jwt.RegisteredClaims
}

// VerifyAccessToken exposes token verification for the bearer-auth path in the
// session middleware (an admin client may send Authorization: Bearer instead of
// the session cookie).
func (h *httpAuthenticator) VerifyAccessToken(raw string) (Identity, error) {
	return h.verifyToken(raw)
}

// verifyToken parses and cryptographically verifies raw (HS256 + expiry +
// audience==site), then returns the identity carried in its claims.
func (h *httpAuthenticator) verifyToken(raw string) (Identity, error) {
	claims := &accessClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %q", t.Header["alg"])
		}
		return h.jwtSecret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithAudience(h.site),
	)
	if err != nil {
		return Identity{}, fmt.Errorf("verify access token: %w", err)
	}
	if claims.Site != h.site {
		return Identity{}, fmt.Errorf("token site claim %q != %q", claims.Site, h.site)
	}
	if want := strings.TrimRight(h.issuer, "/"); want != "" {
		if got := strings.TrimRight(claims.Issuer, "/"); got != want {
			return Identity{}, fmt.Errorf("token issuer %q != %q", claims.Issuer, want)
		}
	}
	return Identity{
		UserID:      claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.Name,
		Roles:       claims.Roles,
	}, nil
}

func (h *httpAuthenticator) Login(ctx context.Context, email, password string) (Identity, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password, "site": h.site})
	resp, err := h.post(ctx, "/internal/login", body)
	if err != nil {
		return Identity{}, ErrUnreachable
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
		var tr tokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
			return Identity{}, ErrUnreachable
		}
		if tr.MFARequired || tr.AccessToken == "" {
			return Identity{}, ErrMFARequired
		}
		id, err := h.verifyToken(tr.AccessToken)
		if err != nil {
			h.log().Error("login: access token verification failed", "err", err)
			return Identity{}, ErrUnreachable
		}
		return id, nil
	case resp.StatusCode == http.StatusUnauthorized:
		return Identity{}, ErrBadCredentials
	case resp.StatusCode == http.StatusForbidden:
		return Identity{}, ErrDisabled
	case resp.StatusCode == http.StatusConflict:
		return Identity{}, ErrMFARequired
	default:
		return Identity{}, ErrUnreachable
	}
}

func (h *httpAuthenticator) Mint(ctx context.Context, userID string) (Identity, error) {
	body, _ := json.Marshal(map[string]string{"user_id": userID, "site": h.site})
	resp, err := h.post(ctx, "/internal/token/mint", body)
	if err != nil {
		return Identity{}, ErrUnreachable
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
		var tr tokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
			return Identity{}, ErrUnreachable
		}
		id, err := h.verifyToken(tr.AccessToken)
		if err != nil {
			h.log().Error("mint: access token verification failed", "err", err)
			return Identity{}, ErrUnreachable
		}
		if id.UserID == "" {
			id.UserID = userID
		}
		return id, nil
	case resp.StatusCode == http.StatusForbidden,
		resp.StatusCode == http.StatusNotFound,
		resp.StatusCode == http.StatusConflict:
		return Identity{}, ErrUserClosed
	default:
		return Identity{}, ErrUnreachable
	}
}

func (h *httpAuthenticator) post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Secret", h.secret)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth %s: %w", path, err)
	}
	return resp, nil
}
