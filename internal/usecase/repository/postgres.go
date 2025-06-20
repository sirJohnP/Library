package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/project/library/internal/entity"
)

var _ BookRepository = (*postgresRepository)(nil)
var _ AuthorRepository = (*postgresRepository)(nil)

const (
	errForeignKeyViolation = "23503"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		db: db,
	}
}

func getError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == errForeignKeyViolation {
		return fmt.Errorf("some authors does not exist: %w", entity.ErrAuthorNotFound)
	}

	return err
}

func addAuthorBooks(ctx context.Context, tx pgx.Tx, bookID string, authorIDs []string) error {
	rows := make([][]any, len(authorIDs))
	for i, authorID := range authorIDs {
		rows[i] = []any{authorID, bookID}
	}

	_, err := tx.Conn().CopyFrom(ctx, pgx.Identifier{"author_book"}, []string{"author_id", "book_id"}, pgx.CopyFromRows(rows))

	return getError(err)
}

func getAuthorsList(authorIDs []sql.NullString) []string {
	result := make([]string, 0)
	if authorIDs[0].Valid {
		for elem := range authorIDs {
			result = append(result, authorIDs[elem].String)
		}
	}
	return result
}

func (p postgresRepository) CreateBook(ctx context.Context, book entity.Book) (resBook entity.Book, txErr error) {
	var (
		tx  pgx.Tx
		err error
	)

	if tx, err = extractTX(ctx); err != nil {
		tx, err = p.db.Begin(ctx)

		if err == nil {
			defer func(cxt context.Context, tx pgx.Tx) {
				if txErr != nil {
					_ = tx.Rollback(ctx)
					return
				}

				_ = tx.Commit(cxt)
			}(ctx, tx)
		}
	}

	if err != nil {
		return entity.Book{}, err
	}

	const queryBook = `INSERT INTO book (name) VALUES ($1) RETURNING id, created_at, updated_at`

	result := entity.Book{
		Name:      book.Name,
		AuthorIDs: book.AuthorIDs,
	}

	if err := tx.QueryRow(ctx, queryBook, result.Name).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt); err != nil {
		return entity.Book{}, err
	}

	if err := addAuthorBooks(ctx, tx, result.ID, result.AuthorIDs); err != nil {
		return entity.Book{}, err
	}

	return result, nil
}

func (p postgresRepository) GetBook(ctx context.Context, bookID string) (entity.Book, error) {
	const query = `SELECT book.id, book.name, book.created_at, book.updated_at, array_agg(author_book.author_id) 
					FROM book LEFT JOIN author_book ON book.id = author_book.book_id 
					WHERE book.id = $1
					GROUP BY book.id, book.name, book.created_at, book.updated_at`

	var result entity.Book
	var authorIDs []sql.NullString
	err := p.db.QueryRow(ctx, query, bookID).Scan(&result.ID, &result.Name, &result.CreatedAt, &result.UpdatedAt, &authorIDs)

	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Book{}, entity.ErrBookNotFound
	}
	if err != nil {
		return entity.Book{}, err
	}
	result.AuthorIDs = getAuthorsList(authorIDs)

	return result, nil
}

func (p postgresRepository) ChangeBookInfo(ctx context.Context, bookID string, newBook entity.Book) (entity.Book, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.Book{}, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	result := entity.Book{
		ID:        bookID,
		Name:      newBook.Name,
		AuthorIDs: newBook.AuthorIDs,
	}

	const query = `UPDATE book SET name = $2 WHERE id = $1`
	_, err = tx.Exec(ctx, query, bookID, result.Name)
	if err != nil {
		return entity.Book{}, err
	}

	const queryCurrentAuthors = `SELECT author_id FROM author_book WHERE book_id = $1`
	rows, err := tx.Query(ctx, queryCurrentAuthors, bookID)
	if err != nil {
		return entity.Book{}, err
	}
	defer rows.Close()
	authorsToDelete := make([]string, 0)
	currentAuthors := make([]string, 0)
	for rows.Next() {
		var authorID string
		if err = rows.Scan(&authorID); err != nil {
			return entity.Book{}, err
		}
		currentAuthors = append(currentAuthors, authorID)
		fl := false
		for _, newAuthorID := range result.AuthorIDs {
			if newAuthorID == authorID {
				fl = true
			}
		}
		if !fl {
			authorsToDelete = append(authorsToDelete, authorID)
		}
	}

	authorsToInsert := make([]string, 0)
	for _, newAuthorID := range result.AuthorIDs {
		fl := true
		for _, authorID := range currentAuthors {
			if newAuthorID == authorID {
				fl = false
				break
			}
		}
		if fl {
			authorsToInsert = append(authorsToInsert, newAuthorID)
		}
	}

	const queryRemoveAuthors = `DELETE FROM author_book WHERE author_id = Any($1)`
	_, err = tx.Exec(ctx, queryRemoveAuthors, authorsToDelete)
	if err != nil {
		return entity.Book{}, err
	}

	if err := addAuthorBooks(ctx, tx, result.ID, authorsToInsert); err != nil {
		return entity.Book{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return entity.Book{}, err
	}

	return result, nil
}

func (p postgresRepository) GetBooksByAuthor(ctx context.Context, authorIDs string) ([]entity.Book, error) {
	const query = `SELECT book.id, book.name, book.created_at, book.updated_at, array_agg(author_book.author_id) 
					FROM book LEFT JOIN author_book ON book.id = author_book.book_id 
					WHERE book.id = ANY(SELECT book_id FROM author_book WHERE author_book.author_id = $1)
					GROUP BY book.id, book.name, book.created_at, book.updated_at`

	rows, err := p.db.Query(ctx, query, authorIDs)
	if err != nil {
		return []entity.Book{}, err
	}
	defer rows.Close()

	var books []entity.Book
	for rows.Next() {
		var book entity.Book
		var authorIDs []sql.NullString
		if err := rows.Scan(&book.ID, &book.Name, &book.CreatedAt, &book.UpdatedAt, &authorIDs); err != nil {
			return []entity.Book{}, err
		}

		book.AuthorIDs = getAuthorsList(authorIDs)

		books = append(books, book)
	}
	return books, nil
}

func (p postgresRepository) CreateAuthor(ctx context.Context, author entity.Author) (resAuthor entity.Author, txErr error) {
	var (
		tx  pgx.Tx
		err error
	)

	if tx, err = extractTX(ctx); err != nil {
		tx, err = p.db.Begin(ctx)

		if err == nil {
			defer func(cxt context.Context, tx pgx.Tx) {
				if txErr != nil {
					_ = tx.Rollback(ctx)
					return
				}

				_ = tx.Commit(cxt)
			}(ctx, tx)
		}
	}

	if err != nil {
		return entity.Author{}, err
	}

	const query = `INSERT INTO author (name) VALUES ($1) RETURNING id`

	result := entity.Author{
		Name: author.Name,
	}

	err = tx.QueryRow(ctx, query, result.Name).Scan(&result.ID)

	if err != nil {
		return entity.Author{}, err
	}

	return result, nil
}

func (p postgresRepository) GetAuthor(ctx context.Context, authorID string) (entity.Author, error) {
	const query = `SELECT id, name FROM author WHERE id = ($1)`

	var author entity.Author
	err := p.db.QueryRow(ctx, query, authorID).Scan(&author.ID, &author.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Author{}, entity.ErrAuthorNotFound
	}

	if err != nil {
		return entity.Author{}, err
	}

	return author, nil
}

func (p postgresRepository) ChangeAuthorInfo(ctx context.Context, id string, newAuthor entity.Author) (entity.Author, error) {
	const query = `UPDATE author SET name = $2 WHERE id = $1`

	result := entity.Author{
		ID:   id,
		Name: newAuthor.Name,
	}
	res, err := p.db.Exec(ctx, query, id, result.Name)

	if err != nil {
		return entity.Author{}, err
	}

	if res.RowsAffected() == 0 {
		return entity.Author{}, entity.ErrAuthorNotFound
	}

	return result, nil
}
