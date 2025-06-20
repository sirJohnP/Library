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

func TestControllerGetBookInfo(t *testing.T) {
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
			name:    "invalid book uid",
			prepare: emptyBookUseCasePrepare,
			book: &library.Book{
				Id:       "some invalid uuid",
				Name:     book.GetName(),
				AuthorId: book.GetAuthorId(),
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name: "book not found",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().GetBook(ctx, book.GetId()).Return(nil, entity.ErrBookNotFound)
			},
			book:         book,
			expectedCode: codes.NotFound,
			noError:      false,
		},
		{
			name: "success",
			prepare: func(mock *mocks.MockBookUseCase) {
				mock.EXPECT().GetBook(ctx, book.GetId()).Return(&library.GetBookInfoResponse{
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

			result, err := data.impl.GetBookInfo(ctx, &library.GetBookInfoRequest{
				Id: tt.book.GetId(),
			})
			if tt.noError {
				require.NoError(t, err)
				require.NotNil(t, result)
				compareBooks(t, book, result.GetBook())
			} else {
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedCode, s.Code())
			}
		})
	}
}
