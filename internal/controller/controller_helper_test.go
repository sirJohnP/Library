package controller

import (
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/usecase/library/mocks"
	"github.com/stretchr/testify/require"
)

type controllerData struct {
	authorUseCase *mocks.MockAuthorUseCase
	bookUseCase   *mocks.MockBookUseCase
	impl          *implementation
}

func emptyBookUseCasePrepare(_ *mocks.MockBookUseCase) {}

func emptyAuthorUseCasePrepare(_ *mocks.MockAuthorUseCase) {}

func compareBooks(t *testing.T, a *library.Book, b *library.Book) {
	t.Helper()
	require.Equal(t, a.GetId(), b.GetId())
	require.Equal(t, a.GetName(), b.GetName())
	require.ElementsMatch(t, a.GetAuthorId(), b.GetAuthorId())
}

func getControllerData(t *testing.T) *controllerData {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthorUseCase := mocks.NewMockAuthorUseCase(ctrl)
	mockBookUseCase := mocks.NewMockBookUseCase(ctrl)

	impl := New(&zap.Logger{}, mockBookUseCase, mockAuthorUseCase)

	return &controllerData{
		authorUseCase: mockAuthorUseCase,
		bookUseCase:   mockBookUseCase,
		impl:          impl,
	}
}
