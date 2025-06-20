package controller

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerChangeAuthorInfo(t *testing.T) {
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
			name:    "invalid author id",
			prepare: emptyAuthorUseCasePrepare,
			author: entity.Author{
				ID:   "some invalid uuid",
				Name: author.Name,
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name:    "invalid author name",
			prepare: emptyAuthorUseCasePrepare,
			author: entity.Author{
				ID:   author.ID,
				Name: "!!!!!",
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name:    "name too short",
			prepare: emptyAuthorUseCasePrepare,
			author: entity.Author{
				ID:   author.ID,
				Name: "",
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name:    "name too long",
			prepare: emptyAuthorUseCasePrepare,
			author: entity.Author{
				ID:   author.ID,
				Name: strings.Repeat("a", 513),
			},
			expectedCode: codes.InvalidArgument,
			noError:      false,
		},
		{
			name: "author does not exist",
			prepare: func(mock *mocks.MockAuthorUseCase) {
				mock.EXPECT().ChangeAuthorInfo(ctx, author.ID, author.Name).Return(entity.ErrAuthorNotFound)
			},
			author:       author,
			expectedCode: codes.NotFound,
			noError:      false,
		},
		{
			name: "success",
			prepare: func(mock *mocks.MockAuthorUseCase) {
				mock.EXPECT().ChangeAuthorInfo(ctx, author.ID, author.Name).Return(nil)
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

			result, err := data.impl.ChangeAuthorInfo(ctx, &library.ChangeAuthorInfoRequest{
				Id:   tt.author.ID,
				Name: tt.author.Name,
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
