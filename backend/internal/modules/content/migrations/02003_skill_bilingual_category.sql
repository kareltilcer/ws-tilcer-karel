-- content module — make skill.category bilingual (category_cs/category_en),
-- matching the bilingual skill name and every other translatable field. Additive
-- so an already-migrated DB upgrades cleanly: the existing `category` becomes
-- `category_cs`; legacy rows get an empty `category_en` and the public site falls
-- back to category_cs until a Czech/English group label is filled in. Renaming a
-- column carries its references in idx_skill_category_order along automatically.

-- +goose Up

-- +goose StatementBegin
ALTER TABLE skill RENAME COLUMN category TO category_cs;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE skill ADD COLUMN category_en TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
ALTER TABLE skill DROP COLUMN category_en;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE skill RENAME COLUMN category_cs TO category;
-- +goose StatementEnd
