ALTER TABLE users ADD COLUMN building_id BIGINT REFERENCES buildings(id) ON DELETE SET NULL;
ALTER TABLE users DROP CONSTRAINT check_role;
ALTER TABLE users ADD CONSTRAINT check_role CHECK (role IN ('superuser', 'admin', 'guard'));

CREATE INDEX idx_users_building_id ON users(building_id);

COMMENT ON COLUMN users.building_id IS 'Building ID for guards/admins. NULL for superuser, required for guard/admin';

