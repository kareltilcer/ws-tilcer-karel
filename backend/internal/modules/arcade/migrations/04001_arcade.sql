-- arcade module — minigame high-scores (PRD §5). Version block 04000.
-- Games (keys + score bounds) are defined in code (games.go), not the DB; this
-- table only stores submitted scores. Leaderboards read top-N by score desc.

-- +goose Up

-- +goose StatementBegin
CREATE TABLE game_score (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    game_key    TEXT NOT NULL,          -- must match a code-defined game
    player_name TEXT NOT NULL,          -- short, sanitized
    score       INTEGER NOT NULL,       -- code-bounded per game
    locale      TEXT,
    ip          TEXT,                   -- anti-abuse triage
    created_at  TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_game_score_board ON game_score (game_key, score DESC);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS game_score;
-- +goose StatementEnd
