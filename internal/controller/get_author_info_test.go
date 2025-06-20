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

func TestControllerGetAuthorInfo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	author := entity.Author{
		ID:   uuid.New().String(),
		Name: "Author1",
	}

	tests := []struct {
		name         string
		prepare      func(*mocks.MockAuthorUseCase)
		author       entity.Author
		expectedCode codes.Code
		noError      bool
	}{
		{
			name:    "invalid uuid",
			prepare: emptyAuthorUseCasePrepare,
			author: entity.Author{
				ID:   "some invalid uuid",
				Name: author.Name,
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name: "author not found",
			prepare: func(mock *mocks.MockAuthorUseCase) {
				mock.EXPECT().GetAuthor(ctx, author.ID).Return(nil, entity.ErrAuthorNotFound)
			},
			author:       author,
			expectedCode: codes.NotFound,
			noError:      false,
		},
		{
			name: "success",
			prepare: func(mock *mocks.MockAuthorUseCase) {
				mock.EXPECT().GetAuthor(ctx, author.ID).Return(&library.GetAuthorInfoResponse{
					Id:   author.ID,
					Name: author.Name,
				}, nil)
			},
			author:       author,
			expectedCode: codes.OK,
			noError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := getControllerData(t)
			tt.prepare(data.authorUseCase)

			result, err := data.impl.GetAuthorInfo(ctx, &library.GetAuthorInfoRequest{
				Id: tt.author.ID,
			})
			if tt.noError {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, author.ID, result.GetId())
				require.Equal(t, author.Name, result.GetName())
			} else {
				s, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedCode, s.Code())
			}
		})
	}
}
