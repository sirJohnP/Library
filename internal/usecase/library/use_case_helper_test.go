package library

import (
	"testing"

	"github.com/project/library/internal/usecase/repository/mocks"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type useCaseData struct {
	impl             *libraryImpl
	authorRepository *mocks.MockAuthorRepository
	bookRepository   *mocks.MockBookRepository
	outboxRepository *mocks.MockOutboxRepository
	transactor       *mocks.MockTransactor
}

func getUseCaseData(t *testing.T) *useCaseData {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthorRepository := mocks.NewMockAuthorRepository(ctrl)
	mockBookRepository := mocks.NewMockBookRepository(ctrl)
	mockOutboxRepository := mocks.NewMockOutboxRepository(ctrl)
	mockTransactor := mocks.NewMockTransactor(ctrl)

	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	impl := New(logger, mockAuthorRepository, mockBookRepository, mockOutboxRepository, mockTransactor)

	return &useCaseData{
		impl:             impl,
		authorRepository: mockAuthorRepository,
		bookRepository:   mockBookRepository,
		outboxRepository: mockOutboxRepository,
		transactor:       mockTransactor,
	}
}
