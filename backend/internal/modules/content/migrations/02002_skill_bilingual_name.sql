-- content module — make skill.name bilingual (name_cs/name_en), matching every
-- other translatable field. Additive so an already-migrated DB upgrades cleanly:
-- the existing `name` becomes `name_cs`; legacy rows get an empty `name_en` and
-- the public site falls back to name_cs until an English label is filled in.

-- +goose Up

-- +goose StatementBegin
ALTER TABLE skill RENAME COLUMN name TO name_cs;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE skill ADD COLUMN name_en TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
ALTER TABLE skill DROP COLUMN name_en;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE skill RENAME COLUMN name_cs TO name;
-- +goose StatementEnd
