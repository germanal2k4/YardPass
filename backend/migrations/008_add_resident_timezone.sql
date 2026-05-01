-- +goose Up
-- IANA timezone name for per-resident local pass/day boundaries (e.g. Europe/Moscow)
ALTER TABLE residents ADD COLUMN IF NOT EXISTS timezone VARCHAR(100);

-- +goose Down
ALTER TABLE residents DROP COLUMN IF EXISTS timezone;
