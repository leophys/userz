-- 000001 Init: UP

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    first_name TEXT,
    last_name TEXT,
    nickname TEXT NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    email TEXT NOT NULL UNIQUE,
    country TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
