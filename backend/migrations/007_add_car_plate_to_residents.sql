-- +goose Up
-- Add car_plate column to residents table for permanent resident car passes
ALTER TABLE residents ADD COLUMN IF NOT EXISTS car_plate VARCHAR(20);

-- +goose Down
ALTER TABLE residents DROP COLUMN IF EXISTS car_plate;
