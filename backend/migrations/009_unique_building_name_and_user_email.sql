-- +goose Up
DROP INDEX IF EXISTS idx_buildings_name;

CREATE UNIQUE INDEX idx_buildings_name_ci ON buildings (lower(btrim(name)));

CREATE UNIQUE INDEX idx_users_email_ci ON users (lower(btrim(email))) WHERE email IS NOT NULL AND btrim(email) <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_users_email_ci;
DROP INDEX IF EXISTS idx_buildings_name_ci;

CREATE INDEX idx_buildings_name ON buildings(name);
