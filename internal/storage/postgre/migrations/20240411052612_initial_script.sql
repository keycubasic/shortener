-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    date_create TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS links (
    uuid int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INT REFERENCES users(user_id),
    short_url VARCHAR(8),
    original_url VARCHAR(100),
    is_deleted BOOLEAN
);

CREATE INDEX IF NOT EXISTS short_url_idx ON links (short_url);

CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON links (original_url);

GRANT ALL PRIVILEGES ON links TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd