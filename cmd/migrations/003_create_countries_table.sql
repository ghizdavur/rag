-- +migrate Up
CREATE TABLE IF NOT EXISTS countries_table (
    id SERIAL PRIMARY KEY,
    country_name VARCHAR(50) UNIQUE NOT NULL,
    country_flag_url VARCHAR(255) NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS countries_table;
