-- Migration 002: Add password_hash to existing users table (if upgrading from stage-04)
-- Run: psql -d go_backend_production -f migrations/002_add_password_hash.sql

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';
