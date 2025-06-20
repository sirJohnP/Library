-- +goose Up
CREATE INDEX index_author_name ON author (name);

-- +goose Down
DROP INDEX IF EXISTS index_author_name;