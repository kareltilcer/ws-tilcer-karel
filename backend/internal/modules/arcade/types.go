package arcade

// ScoreSubmit mirrors openapi.yaml schema ScoreSubmit. `website` is the honeypot.
type ScoreSubmit struct {
	PlayerName     string `json:"player_name"`
	Score          int64  `json:"score"`
	Locale         string `json:"locale"`
	Website        string `json:"website"`         // honeypot
	TurnstileToken string `json:"turnstile_token"` //
}

// ScoreEntry mirrors openapi.yaml schema ScoreEntry (a leaderboard row).
type ScoreEntry struct {
	PlayerName string `json:"player_name"`
	Score      int64  `json:"score"`
	CreatedAt  string `json:"created_at"`
}

// ScoreSubmitResult mirrors openapi.yaml schema ScoreSubmitResult.
type ScoreSubmitResult struct {
	Rank int64        `json:"rank"`
	Top  []ScoreEntry `json:"top"`
}

// AdminScore is a moderation row (carries the id so the admin can delete it).
type AdminScore struct {
	ID         int64  `json:"id"`
	PlayerName string `json:"player_name"`
	Score      int64  `json:"score"`
	CreatedAt  string `json:"created_at"`
}
