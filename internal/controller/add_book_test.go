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

func TestControllerAddBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	book := &library.Book{
		Id:       uuid.New().String(),
		Name:     "Book1",
		AuthorId: []string{uuid.New().String()},
	}

	tests := []struct {
		name         string
		prepare      func(*mocks.MockBookUseCase)
		book         *library.Book
		expectedCode codes.Code
		noError      bool
	}{
		{
			name:    "invalid author id",
			prepare: emptyBookUseCasePrepare,
			book: &library.Book{
				Id:       book.GetId(),
				Name:     book.GetName(),
				AuthorId: []string{"some invalid uuid"},
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name:    "non unique author ids",
			prepare: emptyBookUseCasePrepare,
			book: &library.Book{
				Id:       book.GetId(),
				Name:     book.GetName(),
				AuthorId: []string{book.GetAuthorId()[0], book.GetAuthorId()[0]},
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name: "author does not exist",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().RegisterBook(ctx, book.GetName(), book.GetAuthorId()).Return(nil, entity.ErrAuthorNotFound)
			},
			book:         book,
			expectedCode: codes.NotFound,
			noError:      false,
		},
		{
			name: "success",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().RegisterBook(ctx, book.GetName(), book.GetAuthorId()).Return(&library.AddBookResponse{
					Book: book,
				}, nil)
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

			result, err := data.impl.AddBook(ctx, &library.AddBookRequest{
				Name:      tt.book.GetName(),
				AuthorIds: tt.book.GetAuthorId(),
			})
			if tt.noError {
				require.NoError(t, err)
				compareBooks(t, book, result.GetBook())
			} else {
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedCode, s.Code())
			}
		})
	}
}
