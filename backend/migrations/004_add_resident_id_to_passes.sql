-- +goose Up
ALTER TABLE passes
ADD COLUMN resident_id BIGINT REFERENCES residents(id) ON DELETE SET NULL;

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

-- +goose Down
ALTER TABLE passes DROP CONSTRAINT IF EXISTS passes_resident_id_fkey;
ALTER TABLE passes DROP COLUMN IF EXISTS resident_id;

DROP INDEX IF EXISTS idx_passes_resident_id;
