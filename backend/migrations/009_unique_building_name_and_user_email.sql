-- +goose Up
-- Only building name must be globally unique. Email is intentionally NOT unique:
-- one operator can register several buildings with the same email, each producing
-- its own admin/guard pair tied to that building.
DROP INDEX IF EXISTS idx_buildings_name;

CREATE UNIQUE INDEX idx_buildings_name_ci ON buildings (lower(btrim(name)));

-- +goose Down
DROP INDEX IF EXISTS idx_buildings_name_ci;

CREATE INDEX idx_buildings_name ON buildings(name);
