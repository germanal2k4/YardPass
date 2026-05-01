-- +goose Up
ALTER TABLE buildings
ADD COLUMN apartment_count INTEGER NOT NULL DEFAULT 1;

ALTER TABLE buildings
ADD CONSTRAINT buildings_apartment_count_positive CHECK (apartment_count > 0);

ALTER TABLE users
ADD COLUMN apartment_number INTEGER;

ALTER TABLE users
ADD CONSTRAINT users_apartment_number_positive CHECK (apartment_number IS NULL OR apartment_number > 0);

-- +goose Down
ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_apartment_number_positive;

ALTER TABLE users
DROP COLUMN IF EXISTS apartment_number;

ALTER TABLE buildings
DROP CONSTRAINT IF EXISTS buildings_apartment_count_positive;

ALTER TABLE buildings
DROP COLUMN IF EXISTS apartment_count;
