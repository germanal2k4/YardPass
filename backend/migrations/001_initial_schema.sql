-- YardPass Database Schema
-- This schema should be reviewed and potentially adjusted by the infrastructure team

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Buildings table (no dependencies)
CREATE TABLE buildings (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_buildings_name ON buildings(name);

-- Users table (no dependencies, needed for scan_events)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'guard',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT check_role CHECK (role IN ('guard', 'admin')),
    CONSTRAINT check_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- Apartments table (depends on buildings)
CREATE TABLE apartments (
    id BIGSERIAL PRIMARY KEY,
    building_id BIGINT NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    number VARCHAR(50) NOT NULL,
    floor INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(building_id, number)
);

CREATE INDEX idx_apartments_building_id ON apartments(building_id);
CREATE INDEX idx_apartments_number ON apartments(number);

-- Residents table (depends on apartments)
CREATE TABLE residents (
    id BIGSERIAL PRIMARY KEY,
    apartment_id BIGINT NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    telegram_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(telegram_id)
);

CREATE INDEX idx_residents_apartment_id ON residents(apartment_id);
CREATE INDEX idx_residents_telegram_id ON residents(telegram_id);
CREATE INDEX idx_residents_status ON residents(status);

-- Passes table (depends on apartments)
CREATE TABLE passes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    apartment_id BIGINT NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    car_plate VARCHAR(20) NOT NULL,
    guest_name VARCHAR(255),
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (valid_to > valid_from)
);

CREATE INDEX idx_passes_apartment_id ON passes(apartment_id);
CREATE INDEX idx_passes_status ON passes(status);
CREATE INDEX idx_passes_valid_to ON passes(valid_to);
CREATE INDEX idx_passes_car_plate ON passes(car_plate);
CREATE INDEX idx_passes_created_at ON passes(created_at);

-- Rules table (depends on buildings)
CREATE TABLE rules (
    id BIGSERIAL PRIMARY KEY,
    building_id BIGINT NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    quiet_hours_start VARCHAR(5), -- HH:MM format
    quiet_hours_end VARCHAR(5),   -- HH:MM format
    daily_pass_limit_per_apartment INTEGER NOT NULL DEFAULT 5,
    max_pass_duration_hours INTEGER NOT NULL DEFAULT 24,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(building_id)
);

CREATE INDEX idx_rules_building_id ON rules(building_id);

-- Scan events table (depends on passes and users - MUST be created last)
CREATE TABLE scan_events (
    id BIGSERIAL PRIMARY KEY,
    pass_id UUID NOT NULL REFERENCES passes(id) ON DELETE CASCADE,
    guard_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    scanned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    result VARCHAR(20) NOT NULL,
    reason TEXT,
    meta JSONB,
    CONSTRAINT check_result CHECK (result IN ('valid', 'invalid'))
);

CREATE INDEX idx_scan_events_pass_id ON scan_events(pass_id);
CREATE INDEX idx_scan_events_guard_user_id ON scan_events(guard_user_id);
CREATE INDEX idx_scan_events_scanned_at ON scan_events(scanned_at);
CREATE INDEX idx_scan_events_result ON scan_events(result);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_buildings_updated_at BEFORE UPDATE ON buildings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_apartments_updated_at BEFORE UPDATE ON apartments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_residents_updated_at BEFORE UPDATE ON residents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_passes_updated_at BEFORE UPDATE ON passes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rules_updated_at BEFORE UPDATE ON rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
