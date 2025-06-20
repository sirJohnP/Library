package library

import (
	"context"

	"github.com/project/library/generated/api/library"

	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

//go:generate ../../../bin/mockgen -source=interfaces.go -destination=mocks/usecase_mock.go -package=mocks

type AuthorUseCase interface {
	RegisterAuthor(ctx context.Context, authorName string) (*library.RegisterAuthorResponse, error)
	GetAuthor(ctx context.Context, authorID string) (*library.GetAuthorInfoResponse, error)
	ChangeAuthorInfo(ctx context.Context, authorID string, newName string) error
}

type BookUseCase interface {
	RegisterBook(ctx context.Context, name string, authorIDs []string) (*library.AddBookResponse, error)
	GetBook(ctx context.Context, bookID string) (*library.GetBookInfoResponse, error)
	ChangeBookInfo(ctx context.Context, bookID string, name string, authorIDs []string) error
	GetBooksByAuthor(ctx context.Context, authorID string) ([]*library.Book, error)
}

var _ AuthorUseCase = (*libraryImpl)(nil)
var _ BookUseCase = (*libraryImpl)(nil)

type libraryImpl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	bookRepository   repository.BookRepository
	outboxRepository repository.OutboxRepository
	transactor       repository.Transactor
}

func New(logger *zap.Logger, authorRepository repository.AuthorRepository, bookRepository repository.BookRepository, outboxRepository repository.OutboxRepository, transactor repository.Transactor) *libraryImpl {
	return &libraryImpl{
		logger:           logger,
		authorRepository: authorRepository,
		bookRepository:   bookRepository,
		outboxRepository: outboxRepository,
		transactor:       transactor,
	}
}
