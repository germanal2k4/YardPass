-- Migration: Add resident_id to passes table
-- Date: 2026-01-05
-- This allows tracking which resident created each pass

ALTER TABLE passes 
ADD COLUMN resident_id BIGINT REFERENCES residents(id) ON DELETE SET NULL;

-- Update existing passes to set resident_id based on apartment_id
-- This is a best-effort migration - some passes might remain NULL
UPDATE passes p
SET resident_id = (
    SELECT r.id 
    FROM residents r 
    WHERE r.apartment_id = p.apartment_id 
    LIMIT 1
)
WHERE resident_id IS NULL;

CREATE INDEX idx_passes_resident_id ON passes(resident_id);

COMMENT ON COLUMN passes.resident_id IS 'Resident who created this pass (NULL for legacy passes)';

