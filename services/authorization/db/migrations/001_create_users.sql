-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION trigger_set_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- +goose Down

DROP TRIGGER IF EXISTS set_updated_at ON users;
DROP FUNCTION IF EXISTS trigger_set_timestamp();
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";