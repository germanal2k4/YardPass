-- +goose Up
ALTER TABLE passes
ALTER COLUMN car_plate DROP NOT NULL;

COMMENT ON COLUMN passes.car_plate IS 'Car plate number (NULL for pedestrian guests)';

-- +goose Down
ALTER TABLE passes ALTER COLUMN car_plate SET NOT NULL;
