-- Rollback for 001_initial_schema.sql
-- This script drops all tables, functions, and extensions created in the initial schema
-- ⚠️  WARNING: This will delete ALL data in the database!

-- Drop all triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_rules_updated_at ON rules;
DROP TRIGGER IF EXISTS update_passes_updated_at ON passes;
DROP TRIGGER IF EXISTS update_residents_updated_at ON residents;
DROP TRIGGER IF EXISTS update_apartments_updated_at ON apartments;
DROP TRIGGER IF EXISTS update_buildings_updated_at ON buildings;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop all tables (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS scan_events CASCADE;
DROP TABLE IF EXISTS rules CASCADE;
DROP TABLE IF EXISTS passes CASCADE;
DROP TABLE IF EXISTS residents CASCADE;
DROP TABLE IF EXISTS apartments CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS buildings CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";

-- Note: This completely removes the database schema
-- All data will be lost!

