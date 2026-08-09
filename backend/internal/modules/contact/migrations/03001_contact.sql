-- contact module — the contact-form inbox (PRD §5). Version block 03000.
-- A submission is stored first (status 'new') and then emailed via Resend; the
-- stored row survives an email failure. ip/user_agent are kept for triage.

-- +goose Up

-- +goose StatementBegin
CREATE TABLE contact_message (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL,             -- reply-to
    subject    TEXT,
    message    TEXT NOT NULL,
    locale     TEXT,                      -- cs | en
    ip         TEXT,
    user_agent TEXT,
    status     TEXT NOT NULL DEFAULT 'new',  -- new | read | archived | spam
    created_at TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_contact_status_created ON contact_message (status, created_at DESC);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS contact_message;
-- +goose StatementEnd
