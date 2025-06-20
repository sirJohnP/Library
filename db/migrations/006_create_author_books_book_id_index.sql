-- +goose Up
CREATE INDEX index_author_book_name ON author_book (book_id);

-- +goose Down
DROP INDEX IF EXISTS index_author_book_name;