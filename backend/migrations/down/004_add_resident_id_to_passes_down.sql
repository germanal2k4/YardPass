-- Rollback for 004_add_resident_id_to_passes.sql
-- This script removes the resident_id column from passes table

-- Drop the foreign key constraint and column
ALTER TABLE passes DROP CONSTRAINT IF EXISTS passes_resident_id_fkey;
ALTER TABLE passes DROP COLUMN IF EXISTS resident_id;

-- Drop the index
DROP INDEX IF EXISTS idx_passes_resident_id;

