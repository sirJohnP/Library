package controller

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerUpdateBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	book := entity.Book{
		ID:        uuid.New().String(),
		Name:      "Book1",
		AuthorIDs: []string{uuid.New().String()},
	}

	tests := []struct {
		name         string
		prepare      func(*mocks.MockBookUseCase)
		book         entity.Book
		expectedCode codes.Code
		noError      bool
	}{
		{
			name:    "invalid book uid",
			prepare: emptyBookUseCasePrepare,
			book: entity.Book{
				ID:        "some invalid uuid",
				Name:      book.Name,
				AuthorIDs: book.AuthorIDs,
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name:    "invalid author id",
			prepare: emptyBookUseCasePrepare,
			book: entity.Book{
				ID:        book.ID,
				Name:      book.Name,
				AuthorIDs: []string{"some invalid uuid"},
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name:    "non unique author ids",
			prepare: emptyBookUseCasePrepare,
			book: entity.Book{
				ID:        book.ID,
				Name:      book.Name,
				AuthorIDs: []string{book.AuthorIDs[0], book.AuthorIDs[0]},
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name: "author does not exist",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().ChangeBookInfo(ctx, book.ID, book.Name, book.AuthorIDs).Return(entity.ErrAuthorNotFound)
			},
			book:         book,
			expectedCode: codes.NotFound,
			noError:      false,
		},
		{
			name: "book not found",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().ChangeBookInfo(ctx, book.ID, book.Name, book.AuthorIDs).Return(entity.ErrBookNotFound)
			},
			book:         book,
			expectedCode: codes.NotFound,
			noError:      false,
		},
		{
			name: "success",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().ChangeBookInfo(ctx, book.ID, book.Name, book.AuthorIDs).Return(nil)
			},
			book:         book,
			expectedCode: codes.OK,
			noError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := getControllerData(t)

			tt.prepare(data.bookUseCase)

			result, err := data.impl.UpdateBook(ctx, &library.UpdateBookRequest{
				Id:        tt.book.ID,
				Name:      tt.book.Name,
				AuthorIds: tt.book.AuthorIDs,
			})
			if tt.noError {
				require.NoError(t, err)
				require.NotNil(t, result)
			} else {
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedCode, s.Code())
			}
		})
	}
}
