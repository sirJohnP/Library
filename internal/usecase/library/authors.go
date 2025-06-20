package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"

	"go.uber.org/zap"
)

func (l *libraryImpl) RegisterAuthor(ctx context.Context, authorName string) (*library.RegisterAuthorResponse, error) {
	var author entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var err error
		author, err = l.authorRepository.CreateAuthor(ctx, entity.Author{
			Name: authorName,
		})

		if err != nil {
			l.logger.Error("cannot create author", zap.Error(err))
			return err
		}

		serialized, err := json.Marshal(author)

		if err != nil {
			l.logger.Error("cannot serialize author", zap.Error(err))
			return err
		}

		idempotencyKey := repository.OutboxKindAuthor.String() + "_" + author.ID
		err = l.outboxRepository.SendMessage(ctx, idempotencyKey, repository.OutboxKindAuthor, serialized)

		if err != nil {
			l.logger.Error("cannot send message to outbox", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}

func (l *libraryImpl) GetAuthor(ctx context.Context, authorID string) (*library.GetAuthorInfoResponse, error) {
	author, err := l.authorRepository.GetAuthor(ctx, authorID)

	if err != nil {
		l.logger.Error("cannot get author", zap.Error(err))
		return nil, err
	}

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}

func (l *libraryImpl) ChangeAuthorInfo(ctx context.Context, authorID string, newName string) error {
	_, err := l.authorRepository.ChangeAuthorInfo(ctx, authorID, entity.Author{
		ID:   authorID,
		Name: newName,
	})

	if err != nil {
		l.logger.Error("error during changing author info", zap.Error(err))
		return err
	}

	return nil
}
