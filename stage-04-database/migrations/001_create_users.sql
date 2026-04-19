-- Migration 001: Create users table
-- Run this manually before starting the server:
--   psql -d go_backend_production -f migrations/001_create_users.sql

-- uuid-ossp extension gives us uuid_generate_v4() to auto-generate UUIDs as primary keys.
-- UUIDs are better than auto-increment integers in production:
--   - No sequential guessing (security)
--   - Safe to generate on the client side
--   - Works across distributed systems
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the users table
CREATE TABLE IF NOT EXISTS users (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed some initial data so we have something to query
INSERT INTO users (name, email) VALUES
    ('Akshad Jaiswal', 'akshad@example.com'),
    ('Sid Tiwatne',    'sid@example.com')
ON CONFLICT (email) DO NOTHING;
