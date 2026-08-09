package arcade

import (
	"context"
	"database/sql"
	"strings"
	"unicode"

	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/spam"
)

const (
	maxPlayerNameLen = 32
	defaultTopLimit  = 10
	maxTopLimit      = 100
)

// profanityList is a minimal blocklist for public leaderboard names (FR-12
// "profanity-guarded"). Extend as needed; matching is case-insensitive substring.
var profanityList = []string{"fuck", "shit", "cunt", "bitch"}

// Service is the arcade module's business logic: code-bounded score submission
// against the game registry, leaderboard reads, and admin moderation.
type Service struct {
	db    *sql.DB
	store *Store
	guard *spam.Guard
}

// NewService builds the arcade service over db.
func NewService(db *sql.DB, guard *spam.Guard) *Service {
	return &Service{db: db, store: NewStore(db), guard: guard}
}

// SubmitScore validates the game + score bound + spam guard, inserts the score,
// and returns the submitter's rank plus the current top board. Rate limiting is
// enforced by the handler.
func (s *Service) SubmitScore(ctx context.Context, gameKey string, in ScoreSubmit, ip, userAgent string) (*ScoreSubmitResult, error) {
	game, ok := GameFor(gameKey)
	if !ok {
		return nil, httpx.ErrNotFound("unknown game")
	}
	name, err := sanitizePlayerName(in.PlayerName)
	if err != nil {
		return nil, err
	}
	if in.Score < 0 || in.Score > game.MaxScore {
		return nil, httpx.ErrUnprocessable("score out of range")
	}
	if in.Locale != "" && in.Locale != "cs" && in.Locale != "en" {
		return nil, httpx.ErrUnprocessable("locale must be cs or en")
	}
	// spam guard: honeypot + Turnstile (400)
	if err := s.guard.Check(ctx, in.Website, in.TurnstileToken, ip); err != nil {
		return nil, httpx.ErrSpamRejected("failed spam check")
	}

	if _, err := s.store.Insert(ctx, gameKey, name, in.Score, in.Locale, ip); err != nil {
		return nil, err
	}
	rank, err := s.store.RankOf(ctx, gameKey, in.Score)
	if err != nil {
		return nil, err
	}
	top, err := s.store.Top(ctx, gameKey, defaultTopLimit)
	if err != nil {
		return nil, err
	}
	return &ScoreSubmitResult{Rank: rank, Top: top}, nil
}

// Leaderboard returns the top-N scores for a game (404 for an unknown game).
func (s *Service) Leaderboard(ctx context.Context, gameKey string, limit int) ([]ScoreEntry, error) {
	if _, ok := GameFor(gameKey); !ok {
		return nil, httpx.ErrNotFound("unknown game")
	}
	if limit < 1 {
		limit = defaultTopLimit
	}
	if limit > maxTopLimit {
		limit = maxTopLimit
	}
	return s.store.Top(ctx, gameKey, limit)
}

// ListScoresAdmin returns the top scores for a game with their ids (moderation).
func (s *Service) ListScoresAdmin(ctx context.Context, gameKey string, limit int) ([]AdminScore, error) {
	if _, ok := GameFor(gameKey); !ok {
		return nil, httpx.ErrNotFound("unknown game")
	}
	if limit < 1 || limit > maxTopLimit {
		limit = 50
	}
	return s.store.TopWithIDs(ctx, gameKey, limit)
}

// DeleteScore removes a single score (admin moderation).
func (s *Service) DeleteScore(ctx context.Context, id int64) error {
	n, err := s.store.DeleteScore(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.ErrNotFound("score not found")
	}
	return nil
}

// ResetBoard clears a game's leaderboard (404 for an unknown game).
func (s *Service) ResetBoard(ctx context.Context, gameKey string) error {
	if _, ok := GameFor(gameKey); !ok {
		return httpx.ErrNotFound("unknown game")
	}
	return s.store.ResetBoard(ctx, gameKey)
}

// sanitizePlayerName trims, strips control characters, collapses whitespace,
// caps length, and rejects empty or profane names.
func sanitizePlayerName(raw string) (string, error) {
	var b strings.Builder
	lastSpace := false
	for _, r := range strings.TrimSpace(raw) {
		if unicode.IsControl(r) {
			continue
		}
		if unicode.IsSpace(r) {
			if lastSpace {
				continue
			}
			lastSpace = true
			b.WriteRune(' ')
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	name := strings.TrimSpace(b.String())
	if name == "" {
		return "", httpx.ErrUnprocessable("player_name is required")
	}
	if r := []rune(name); len(r) > maxPlayerNameLen {
		name = strings.TrimSpace(string(r[:maxPlayerNameLen]))
	}
	low := strings.ToLower(name)
	for _, bad := range profanityList {
		if strings.Contains(low, bad) {
			return "", httpx.ErrUnprocessable("player_name not allowed")
		}
	}
	return name, nil
}
