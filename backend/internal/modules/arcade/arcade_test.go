package arcade

import (
	"context"
	"errors"
	"testing"

	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/spam"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/testsupport"
)

func apiStatus(err error) int {
	var ae *httpx.APIError
	if errors.As(err, &ae) {
		return ae.Status
	}
	return 0
}

func newSvc(t *testing.T) *Service {
	t.Helper()
	db := testsupport.NewDB(t,
		registry.MigrationSource{Name: "platform", FS: appdb.MigrationsFS},
		registry.MigrationSource{Name: "arcade", FS: MigrationsFS},
	)
	return NewService(db, spam.NewGuard("")) // Turnstile disabled; honeypot enforced
}

func TestSubmitScoreRankAndTop(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()
	game := "catch-the-scoop"

	submit := func(name string, score int64) *ScoreSubmitResult {
		res, err := s.SubmitScore(ctx, game, ScoreSubmit{PlayerName: name, Score: score}, "1.2.3.4", "")
		if err != nil {
			t.Fatalf("SubmitScore(%s,%d): %v", name, score, err)
		}
		return res
	}
	submit("Alice", 100)
	submit("Bob", 300)
	res := submit("Cara", 200) // should rank #2 (behind Bob 300)
	if res.Rank != 2 {
		t.Fatalf("Cara rank = %d, want 2", res.Rank)
	}
	if len(res.Top) != 3 || res.Top[0].PlayerName != "Bob" || res.Top[0].Score != 300 {
		t.Fatalf("top board wrong: %+v", res.Top)
	}
}

func TestSubmitUnknownGame(t *testing.T) {
	s := newSvc(t)
	if _, err := s.SubmitScore(context.Background(), "no-such-game", ScoreSubmit{PlayerName: "X", Score: 1}, "1.2.3.4", ""); apiStatus(err) != 404 {
		t.Fatalf("unknown game: want 404, got err=%v", err)
	}
}

func TestSubmitScoreOutOfRange(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()
	if _, err := s.SubmitScore(ctx, "catch-the-scoop", ScoreSubmit{PlayerName: "X", Score: 999999999}, "1.2.3.4", ""); apiStatus(err) != 422 {
		t.Fatalf("over max: want 422, got err=%v", err)
	}
	if _, err := s.SubmitScore(ctx, "catch-the-scoop", ScoreSubmit{PlayerName: "X", Score: -1}, "1.2.3.4", ""); apiStatus(err) != 422 {
		t.Fatalf("negative: want 422, got err=%v", err)
	}
}

func TestSubmitHoneypotRejected(t *testing.T) {
	s := newSvc(t)
	if _, err := s.SubmitScore(context.Background(), "catch-the-scoop", ScoreSubmit{PlayerName: "X", Score: 1, Website: "bot"}, "1.2.3.4", ""); apiStatus(err) != 400 {
		t.Fatalf("honeypot: want 400, got err=%v", err)
	}
}

func TestSubmitEmptyAndProfaneName(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()
	if _, err := s.SubmitScore(ctx, "catch-the-scoop", ScoreSubmit{PlayerName: "   ", Score: 1}, "1.2.3.4", ""); apiStatus(err) != 422 {
		t.Fatalf("empty name: want 422, got err=%v", err)
	}
	if _, err := s.SubmitScore(ctx, "catch-the-scoop", ScoreSubmit{PlayerName: "you shit", Score: 1}, "1.2.3.4", ""); apiStatus(err) != 422 {
		t.Fatalf("profane name: want 422, got err=%v", err)
	}
}

func TestLeaderboardAndModeration(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()
	game := "scoop-match"
	for i := int64(1); i <= 3; i++ {
		if _, err := s.SubmitScore(ctx, game, ScoreSubmit{PlayerName: "P", Score: i * 10}, "1.2.3.4", ""); err != nil {
			t.Fatalf("submit: %v", err)
		}
	}
	// unknown game leaderboard → 404
	if _, err := s.Leaderboard(ctx, "nope", 10); apiStatus(err) != 404 {
		t.Fatalf("unknown leaderboard: want 404, got err=%v", err)
	}
	board, err := s.Leaderboard(ctx, game, 10)
	if err != nil || len(board) != 3 || board[0].Score != 30 {
		t.Fatalf("leaderboard wrong: %+v (err=%v)", board, err)
	}
	// reset clears the board
	if err := s.ResetBoard(ctx, game); err != nil {
		t.Fatalf("reset: %v", err)
	}
	board, _ = s.Leaderboard(ctx, game, 10)
	if len(board) != 0 {
		t.Fatalf("after reset board = %d, want 0", len(board))
	}
	// delete unknown score → 404
	if err := s.DeleteScore(ctx, 99999); apiStatus(err) != 404 {
		t.Fatalf("delete unknown: want 404, got err=%v", err)
	}
}
