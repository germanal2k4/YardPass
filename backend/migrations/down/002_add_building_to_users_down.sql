-- Rollback for 002_add_building_to_users.sql
-- This script removes building_id from users and reverts role constraint

-- Drop the index
DROP INDEX IF EXISTS idx_users_building_id;

-- Remove the column
ALTER TABLE users DROP COLUMN IF EXISTS building_id;

-- Revert role constraint to original values
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_role;
ALTER TABLE users ADD CONSTRAINT check_role CHECK (role IN ('guard', 'admin'));

