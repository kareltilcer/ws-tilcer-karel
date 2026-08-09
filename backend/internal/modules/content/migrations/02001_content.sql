-- content module — all content + media tables (PRD §5). Version block 02000.
-- The `media` table lives here (rather than in the media module) because the
-- content tables carry FKs into it and a single file keeps the create order
-- correct: media (no deps) → category → project → project_media → link →
-- page_section → skill → business_info. The media module (M2) reuses this table.
--
-- Every translatable field is a *_cs/*_en pair. Timestamps are RFC3339 UTC TEXT.
-- FK behaviour (PRD §5): project.category_id ON DELETE RESTRICT (409 in-use);
-- project_media CASCADE; cover_media_id / page_section.media_id SET NULL.

-- +goose Up

-- +goose StatementBegin
CREATE TABLE media (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    s3_key     TEXT NOT NULL,
    public_url TEXT NOT NULL,
    mime       TEXT NOT NULL,
    width      INTEGER,
    height     INTEGER,
    size_bytes INTEGER NOT NULL,
    alt_cs     TEXT,
    alt_en     TEXT,
    created_at TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE category (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    slug       TEXT NOT NULL UNIQUE,
    name_cs    TEXT NOT NULL,
    name_en    TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    visible    INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_category_visible_order ON category (visible, sort_order);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE project (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    slug           TEXT NOT NULL UNIQUE,
    category_id    INTEGER NOT NULL REFERENCES category(id) ON DELETE RESTRICT,
    title_cs       TEXT NOT NULL,
    title_en       TEXT NOT NULL,
    summary_cs     TEXT,
    summary_en     TEXT,
    body_cs        TEXT,
    body_en        TEXT,
    cover_media_id INTEGER REFERENCES media(id) ON DELETE SET NULL,
    links          TEXT,               -- JSON array of {label,url,type}
    project_date   TEXT,
    featured       INTEGER NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'draft',  -- draft | published
    sort_order     INTEGER NOT NULL DEFAULT 0,
    created_at     TEXT NOT NULL,
    updated_at     TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_project_status_order ON project (status, sort_order);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_project_category ON project (category_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE project_media (
    project_id INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    media_id   INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    caption_cs TEXT,
    caption_en TEXT,
    PRIMARY KEY (project_id, media_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_project_media_order ON project_media (project_id, sort_order);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE link (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    label_cs    TEXT NOT NULL,
    label_en    TEXT NOT NULL,
    url         TEXT NOT NULL,
    icon        TEXT,
    visible     INTEGER NOT NULL DEFAULT 1,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    click_count INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_link_visible_order ON link (visible, sort_order);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE page_section (
    key        TEXT PRIMARY KEY,       -- e.g. about, hero
    heading_cs TEXT,
    heading_en TEXT,
    body_cs    TEXT,
    body_en    TEXT,
    media_id   INTEGER REFERENCES media(id) ON DELETE SET NULL,
    updated_at TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE skill (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    category   TEXT NOT NULL,          -- grouping (languages/tools/craft)
    icon       TEXT,
    level      INTEGER,                -- 1..5, optional
    visible    INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_skill_category_order ON skill (category, sort_order);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE business_info (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    legal_name       TEXT NOT NULL DEFAULT '',
    ico              TEXT,
    dic              TEXT,             -- empty (neplátce); renders only if set
    address_lines    TEXT,
    city             TEXT,
    zip              TEXT,
    country          TEXT,
    register_note_cs TEXT,
    register_note_en TEXT,
    contact_email    TEXT,
    contact_phone    TEXT,
    data_note_cs     TEXT,
    data_note_en     TEXT,
    updated_at       TEXT NOT NULL
);
-- +goose StatementEnd

-- Seeds: the business_info singleton and the two editable page sections. No
-- categories are seeded (Karel adds them in the CMS). INSERT OR IGNORE keeps a
-- Litestream-restored DB from double-seeding.
-- +goose StatementBegin
INSERT OR IGNORE INTO business_info (id, legal_name, updated_at)
VALUES (1, '', strftime('%Y-%m-%dT%H:%M:%SZ','now'));
-- +goose StatementEnd
-- +goose StatementBegin
INSERT OR IGNORE INTO page_section (key, updated_at)
VALUES ('about', strftime('%Y-%m-%dT%H:%M:%SZ','now')),
       ('hero',  strftime('%Y-%m-%dT%H:%M:%SZ','now'));
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS business_info;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS skill;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS page_section;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS link;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS project_media;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS project;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS category;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS media;
-- +goose StatementEnd
