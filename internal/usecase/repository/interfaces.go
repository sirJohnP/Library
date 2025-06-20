package repository

import (
	"context"
	"time"

	"github.com/project/library/internal/entity"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=mocks/repository_mock.go -package=mocks

type AuthorRepository interface {
	CreateAuthor(ctx context.Context, author entity.Author) (entity.Author, error)
	GetAuthor(ctx context.Context, id string) (entity.Author, error)
	ChangeAuthorInfo(ctx context.Context, id string, newAuthor entity.Author) (entity.Author, error)
}

type BookRepository interface {
	CreateBook(ctx context.Context, book entity.Book) (entity.Book, error)
	GetBook(ctx context.Context, id string) (entity.Book, error)
	ChangeBookInfo(ctx context.Context, id string, newBook entity.Book) (entity.Book, error)
	GetBooksByAuthor(ctx context.Context, authorID string) ([]entity.Book, error)
}

type OutboxRepository interface {
	SendMessage(ctx context.Context, idempotencyKey string, kind OutboxKind, message []byte) error
	GetMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]OutboxData, error)
	MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error
}

type OutboxKind int

type OutboxData struct {
	IdempotencyKey string
	Kind           OutboxKind
	RawData        []byte
}

const (
	OutboxKindUndefined OutboxKind = iota
	OutboxKindBook
	OutboxKindAuthor
)

func (o OutboxKind) String() string {
	switch o {
	case OutboxKindBook:
		return "book"
	case OutboxKindAuthor:
		return "author"
	default:
		return "undefined"
	}
}
