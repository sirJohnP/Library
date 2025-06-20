package library

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBookUseCase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	book := entity.Book{
		ID:        uuid.New().String(),
		Name:      "Book1",
		AuthorIDs: []string{uuid.New().String(), uuid.New().String()},
	}

	tests := []struct {
		testName            string
		prepare             func(*useCaseData)
		apply               func(*useCaseData) (*library.Book, error)
		requireNonNilResult bool
		wantErr             error
	}{
		{
			testName: "createBook successfully",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().CreateBook(ctx, gomock.AssignableToTypeOf(book)).Return(book, nil)
				data.outboxRepository.EXPECT().SendMessage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				data.transactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, x func(context.Context) error) error {
					return x(ctx)
				})
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				resp, err := data.impl.RegisterBook(ctx, book.Name, book.AuthorIDs)
				return resp.GetBook(), err
			},
			requireNonNilResult: true,
			wantErr:             nil,
		},
		{
			testName: "createBook with no existing authors",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().CreateBook(ctx, gomock.AssignableToTypeOf(book)).Return(book, entity.ErrAuthorNotFound)
				data.transactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, x func(context.Context) error) error {
					return x(ctx)
				})
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				resp, err := data.impl.RegisterBook(ctx, book.Name, book.AuthorIDs)
				return resp.GetBook(), err
			},
			requireNonNilResult: false,
			wantErr:             entity.ErrAuthorNotFound,
		},
		{
			testName: "getBook successfully",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().GetBook(ctx, book.ID).Return(book, nil)
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				resp, err := data.impl.GetBook(ctx, book.ID)
				return resp.GetBook(), err
			},
			requireNonNilResult: true,
			wantErr:             nil,
		},
		{
			testName: "getBook book not found",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().GetBook(ctx, book.ID).Return(entity.Book{}, entity.ErrBookNotFound)
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				resp, err := data.impl.GetBook(ctx, book.ID)
				return resp.GetBook(), err
			},
			requireNonNilResult: false,
			wantErr:             entity.ErrBookNotFound,
		},
		{
			testName: "changeBook successfully",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().ChangeBookInfo(ctx, book.ID, book).Return(book, nil)
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				return nil, data.impl.ChangeBookInfo(ctx, book.ID, book.Name, book.AuthorIDs)
			},
			requireNonNilResult: false,
			wantErr:             nil,
		},
		{
			testName: "changeBook book not found",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().ChangeBookInfo(ctx, book.ID, book).Return(entity.Book{}, entity.ErrBookNotFound)
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				return nil, data.impl.ChangeBookInfo(ctx, book.ID, book.Name, book.AuthorIDs)
			},
			requireNonNilResult: false,
			wantErr:             entity.ErrBookNotFound,
		},
		{
			testName: "changeBook with non existing authors",
			prepare: func(data *useCaseData) {
				data.bookRepository.EXPECT().ChangeBookInfo(ctx, book.ID, book).Return(entity.Book{}, entity.ErrAuthorNotFound)
			},
			apply: func(data *useCaseData) (*library.Book, error) {
				return nil, data.impl.ChangeBookInfo(ctx, book.ID, book.Name, book.AuthorIDs)
			},
			requireNonNilResult: false,
			wantErr:             entity.ErrAuthorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
			tData := getUseCaseData(t)

			tt.prepare(tData)

			result, err := tt.apply(tData)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				if tt.requireNonNilResult {
					require.Equal(t, book.ID, result.GetId())
					require.Equal(t, book.Name, result.GetName())
					require.ElementsMatch(t, book.AuthorIDs, result.GetAuthorId())
				} else {
					require.Nil(t, result)
				}
			}
		})
	}
}

func TestUseCaseGetBooksByAuthor(t *testing.T) {
	t.Parallel()
	data := getUseCaseData(t)

	ctx := context.Background()
	author1 := entity.Author{
		ID:   uuid.New().String(),
		Name: "Author1",
	}
	books := make([]entity.Book, 10)
	bookIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		books[i] = entity.Book{
			ID:        uuid.New().String(),
			Name:      "Book" + strconv.Itoa(i),
			AuthorIDs: []string{author1.ID},
		}
		bookIDs[i] = books[i].ID
	}

	t.Run("get books by author successfully", func(t *testing.T) {
		t.Parallel()
		data.bookRepository.EXPECT().GetBooksByAuthor(ctx, author1.ID).Return(books, nil)

		result, err := data.impl.GetBooksByAuthor(ctx, author1.ID)
		require.NoError(t, err)

		resultIDs := make([]string, 10)
		for i := 0; i < 10; i++ {
			resultIDs[i] = result[i].GetId()
		}

		require.ElementsMatch(t, bookIDs, resultIDs)
	})
	t.Run("get books by non existing author", func(t *testing.T) {
		t.Parallel()
		data.bookRepository.EXPECT().GetBooksByAuthor(ctx, author1.ID).Return(nil, entity.ErrAuthorNotFound)

		_, err := data.impl.GetBooksByAuthor(ctx, author1.ID)
		require.Equal(t, entity.ErrAuthorNotFound, err)
	})
}
