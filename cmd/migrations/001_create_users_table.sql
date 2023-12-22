-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL COLLATE "C" -- Add COLLATE "C" for case-insensitive searches
                      CHECK (username ~* '^[a-z0-9_]+$'), -- Optional: Enforce lowercase alphanumeric characters and underscores
    passwd VARCHAR(255) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ DEFAULT current_timestamp,
    deleted_at TIMESTAMPTZ
);


-- +migrate Down
DROP TABLE IF EXISTS users;
