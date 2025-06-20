package controller

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerConvertError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		status codes.Code
	}{
		{
			name:   "book not found error",
			err:    entity.ErrBookNotFound,
			status: codes.NotFound,
		},
		{
			name:   "author not found error",
			err:    entity.ErrAuthorNotFound,
			status: codes.NotFound,
		},
		{
			name:   "unknown error",
			err:    errors.New("unknown error"),
			status: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
			mockBookUseCase := mocks.NewMockBookUseCase(ctrl)

			impl := New(&zap.Logger{}, mockBookUseCase, mockAuthorUseCase)

			err := impl.convertError(tt.err)
			s, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.status, s.Code())
		})
	}
}
