package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/generated/api/library"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"

	"go.uber.org/zap"
)

func convertBookToResponse(book entity.Book) *library.Book {
	return &library.Book{
		Id:        book.ID,
		Name:      book.Name,
		AuthorId:  book.AuthorIDs,
		CreatedAt: timestamppb.New(book.CreatedAt),
		UpdatedAt: timestamppb.New(book.UpdatedAt),
	}
}

func (l *libraryImpl) RegisterBook(ctx context.Context, name string, authorIDs []string) (*library.AddBookResponse, error) {
	var book entity.Book

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var err error
		book, err = l.bookRepository.CreateBook(ctx, entity.Book{
			Name:      name,
			AuthorIDs: authorIDs,
		})

		if err != nil {
			l.logger.Error("cannot create book", zap.Error(err))
			return err
		}

		serialized, err := json.Marshal(book)

		if err != nil {
			l.logger.Error("cannot serialize book", zap.Error(err))
			return err
		}

		idempotencyKey := repository.OutboxKindBook.String() + "_" + book.ID
		err = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindBook, serialized)

		if err != nil {
			l.logger.Error("cannot send message to outbox", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &library.AddBookResponse{
		Book: convertBookToResponse(book),
	}, nil
}

func (l *libraryImpl) GetBook(ctx context.Context, bookID string) (*library.GetBookInfoResponse, error) {
	book, err := l.bookRepository.GetBook(ctx, bookID)

	if err != nil {
		l.logger.Error("cannot get book", zap.Error(err))
		return nil, err
	}

	return &library.GetBookInfoResponse{
		Book: convertBookToResponse(book),
	}, nil
}

func (l *libraryImpl) ChangeBookInfo(ctx context.Context, bookID string, name string, authorIDs []string) error {
	_, err := l.bookRepository.ChangeBookInfo(ctx, bookID, entity.Book{
		ID:        bookID,
		Name:      name,
		AuthorIDs: authorIDs,
	})

	if err != nil {
		l.logger.Error("cannot change book info", zap.Error(err))
		return err
	}

	return nil
}

func (l *libraryImpl) GetBooksByAuthor(ctx context.Context, authorID string) ([]*library.Book, error) {
	books, err := l.bookRepository.GetBooksByAuthor(ctx, authorID)

	if err != nil {
		l.logger.Error("cannot get author books", zap.Error(err))
		return nil, err
	}

	res := make([]*library.Book, len(books))
	for i, book := range books {
		res[i] = convertBookToResponse(book)
	}

	return res, nil
}
