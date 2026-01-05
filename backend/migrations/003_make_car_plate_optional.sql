-- Migration: Make car_plate optional for pedestrian guests
-- Date: 2026-01-05

ALTER TABLE passes 
ALTER COLUMN car_plate DROP NOT NULL;

COMMENT ON COLUMN passes.car_plate IS 'Car plate number (NULL for pedestrian guests)';

