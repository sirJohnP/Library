package library

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthorUseCase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	author := entity.Author{
		ID:   uuid.New().String(),
		Name: "Author1",
	}

	tests := []struct {
		testName       string
		prepare        func(*useCaseData)
		apply          func(*useCaseData) (entity.Author, error)
		returnedAuthor entity.Author
		wantErr        error
	}{
		{
			testName: "create author successfully",
			prepare: func(data *useCaseData) {
				data.authorRepository.EXPECT().CreateAuthor(ctx, gomock.AssignableToTypeOf(author)).Return(author, nil)
				data.outboxRepository.EXPECT().SendMessage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				data.transactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, x func(context.Context) error) error {
					return x(ctx)
				})
			},
			apply: func(data *useCaseData) (entity.Author, error) {
				resp, err := data.impl.RegisterAuthor(ctx, author.Name)
				return entity.Author{
					ID: resp.GetId(),
				}, err
			},
			returnedAuthor: entity.Author{
				ID: author.ID,
			},
			wantErr: nil,
		},
		{
			testName: "getAuthor successfully",
			prepare: func(data *useCaseData) {
				data.authorRepository.EXPECT().GetAuthor(ctx, author.ID).Return(author, nil)
			},
			apply: func(data *useCaseData) (entity.Author, error) {
				resp, err := data.impl.GetAuthor(ctx, author.ID)
				return entity.Author{
					ID:   resp.GetId(),
					Name: resp.GetName(),
				}, err
			},
			returnedAuthor: author,
			wantErr:        nil,
		},
		{
			testName: "getAuthor author not found",
			prepare: func(data *useCaseData) {
				data.authorRepository.EXPECT().GetAuthor(ctx, author.ID).Return(entity.Author{}, entity.ErrAuthorNotFound)
			},
			apply: func(data *useCaseData) (entity.Author, error) {
				resp, err := data.impl.GetAuthor(ctx, author.ID)
				return entity.Author{
					ID:   resp.GetId(),
					Name: resp.GetName(),
				}, err
			},
			returnedAuthor: entity.Author{},
			wantErr:        entity.ErrAuthorNotFound,
		},
		{
			testName: "changeAuthorInfo successfully",
			prepare: func(data *useCaseData) {
				data.authorRepository.EXPECT().ChangeAuthorInfo(ctx, author.ID, author).Return(author, nil)
			},
			apply: func(data *useCaseData) (entity.Author, error) {
				err := data.impl.ChangeAuthorInfo(ctx, author.ID, author.Name)
				return entity.Author{}, err
			},
			returnedAuthor: entity.Author{},
			wantErr:        nil,
		},
		{
			testName: "changeAuthorInfo author not found",
			prepare: func(data *useCaseData) {
				data.authorRepository.EXPECT().ChangeAuthorInfo(ctx, author.ID, author).Return(entity.Author{}, entity.ErrAuthorNotFound)
			},
			apply: func(data *useCaseData) (entity.Author, error) {
				err := data.impl.ChangeAuthorInfo(ctx, author.ID, author.Name)
				return entity.Author{}, err
			},
			returnedAuthor: entity.Author{},
			wantErr:        entity.ErrAuthorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
			data := getUseCaseData(t)

			tt.prepare(data)

			result, err := tt.apply(data)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.returnedAuthor.ID, result.ID)
				require.Equal(t, tt.returnedAuthor.Name, result.Name)
			}
		})
	}
}
