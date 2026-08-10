package arcade

import (
	"context"
	"database/sql"
	"time"
)

// Store is the arcade module's SQL access layer over game_score.
type Store struct{ db *sql.DB }

// NewStore returns an arcade store over db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

const tsFormat = "2006-01-02T15:04:05.000000Z07:00"

func nowUTC() string { return time.Now().UTC().Format(tsFormat) }

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Insert stores a score and returns its id.
func (s *Store) Insert(ctx context.Context, gameKey, playerName string, score int64, locale, ip string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO game_score (game_key, player_name, score, locale, ip, created_at)
		 VALUES (?,?,?,?,?,?)`,
		gameKey, playerName, score, nullStr(locale), nullStr(ip), nowUTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Top returns the top-N scores for a game, highest first (ties broken by earliest
// submission).
func (s *Store) Top(ctx context.Context, gameKey string, limit int) ([]ScoreEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT player_name, score, created_at FROM game_score
		 WHERE game_key = ? ORDER BY score DESC, created_at ASC, id ASC LIMIT ?`, gameKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ScoreEntry{}
	for rows.Next() {
		var e ScoreEntry
		if err := rows.Scan(&e.PlayerName, &e.Score, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// TopWithIDs returns the top-N scores for a game including their ids (admin
// moderation).
func (s *Store) TopWithIDs(ctx context.Context, gameKey string, limit int) ([]AdminScore, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, player_name, score, created_at FROM game_score
		 WHERE game_key = ? ORDER BY score DESC, created_at ASC, id ASC LIMIT ?`, gameKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AdminScore{}
	for rows.Next() {
		var a AdminScore
		if err := rows.Scan(&a.ID, &a.PlayerName, &a.Score, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// RankOf returns the 1-based rank of score on a game's board: the number of
// strictly-higher scores plus one.
func (s *Store) RankOf(ctx context.Context, gameKey string, score int64) (int64, error) {
	var higher int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_score WHERE game_key = ? AND score > ?`, gameKey, score).Scan(&higher)
	if err != nil {
		return 0, err
	}
	return higher + 1, nil
}

// DeleteScore removes a single score. Returns rows affected.
func (s *Store) DeleteScore(ctx context.Context, id int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM game_score WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ResetBoard clears every score for a game.
func (s *Store) ResetBoard(ctx context.Context, gameKey string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM game_score WHERE game_key = ?`, gameKey)
	return err
}
