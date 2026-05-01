-- +goose Up
INSERT INTO buildings (id, name, address, created_at, updated_at)
VALUES (1, 'Тестовый корпус', 'ул. Тестовая, д. 1', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO apartments (id, building_id, number, floor, created_at, updated_at)
VALUES (1, 1, '101', 1, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (username, email, password_hash, role, building_id, status, created_at, updated_at)
VALUES (
    'superadmin',
    'superadmin@yardpass.local',
    '$2a$10$Nm8hpHW59yaS2uaI8cKTBedTHzyz2HUfJNjq8QskHB7DvCmqtSTGW',
    'superuser',
    NULL,
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, email, password_hash, role, building_id, status, created_at, updated_at)
VALUES (
    'admin',
    'admin@yardpass.local',
    '$2a$10$Nm8hpHW59yaS2uaI8cKTBedTHzyz2HUfJNjq8QskHB7DvCmqtSTGW',
    'admin',
    1,
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, email, password_hash, role, building_id, status, created_at, updated_at)
VALUES (
    'guard',
    'guard@yardpass.local',
    '$2a$10$9PbhEZlOJaLQDa1S/wnLKu7.vDJKJN3VVXizbCTjsN3Dzgr9hqdIW',
    'guard',
    1,
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (username) DO NOTHING;

INSERT INTO rules (building_id, quiet_hours_start, quiet_hours_end, daily_pass_limit_per_apartment, max_pass_duration_hours, created_at, updated_at)
VALUES (1, '22:00', '08:00', 5, 24, NOW(), NOW())
ON CONFLICT (building_id) DO NOTHING;

COMMENT ON TABLE users IS 'Test accounts: superadmin/admin123, admin/admin123, guard/guard123';

-- +goose Down
DELETE FROM users WHERE username IN ('superadmin', 'admin', 'guard');
DELETE FROM rules WHERE building_id = 1;
DELETE FROM buildings WHERE id = 1;
COMMENT ON TABLE users IS NULL;
