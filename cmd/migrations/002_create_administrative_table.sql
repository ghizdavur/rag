-- +migrate Up
CREATE TABLE IF NOT EXISTS administrative_table (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    user_role VARCHAR(255) NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS administrative_table;
