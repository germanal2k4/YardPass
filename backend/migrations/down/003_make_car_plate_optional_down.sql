-- Rollback for 003_make_car_plate_optional.sql
-- This script makes car_plate required again in passes table

-- Note: This will fail if there are any passes with NULL car_plate
-- You may need to update or delete such records first

-- Make car_plate NOT NULL again
ALTER TABLE passes ALTER COLUMN car_plate SET NOT NULL;

-- Note: This assumes all passes have car_plate values
-- If migration 003 was applied and NULL values exist, run this first:
-- UPDATE passes SET car_plate = 'UNKNOWN' WHERE car_plate IS NULL;

