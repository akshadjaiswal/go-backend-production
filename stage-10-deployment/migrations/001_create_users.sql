-- Migration 001: Create users table
--
-- In Docker, this file is automatically executed by the PostgreSQL container
-- on first startup. docker-compose.yml mounts ./migrations to
-- /docker-entrypoint-initdb.d/ — Postgres runs all .sql files there
-- in alphabetical order when it initialises a fresh data volume.
--
-- This means: no manual migration step! Just `docker compose up` and
-- the schema is created automatically.
--
-- Note: initdb.d only runs on first start (when the data volume is empty).
-- If you change the schema, you need to `docker compose down -v` to wipe
-- the volume and let it re-initialise.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
