package outbox

import (
	"context"
	"testing"

	"github.com/project/library/config"
	"github.com/project/library/internal/usecase/repository"
	"github.com/project/library/internal/usecase/repository/mocks"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type outboxData struct {
	outboxRepository *mocks.MockOutboxRepository
	transactor       *mocks.MockTransactor
	successKeys      []string
}

func (o *outboxData) getImpl(t *testing.T, handler GlobalHandler) *outboxImpl {
	t.Helper()
	cfg := &config.Config{
		GRPC: config.GRPC{},
		PG:   config.PG{},
		Outbox: config.Outbox{
			Enabled: true,
		},
	}
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	return New(logger, o.outboxRepository, handler, cfg, o.transactor)
}

func (o *outboxData) prepareDefaultTransactor() {
	o.transactor.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().
		DoAndReturn(func(ctx context.Context, x func(context.Context) error) error {
			return x(ctx)
		})
}

func (o *outboxData) prepareDefaultOutboxRepository(messages []repository.OutboxData) {
	leftTasks := make([]repository.OutboxData, 0)
	for _, message := range messages {
		fl := true
		for i := 0; i < len(o.successKeys); i++ {
			if o.successKeys[i] == message.IdempotencyKey {
				fl = false
			}
		}
		if fl {
			leftTasks = append(leftTasks, message)
		}
	}
	o.outboxRepository.EXPECT().GetMessages(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(leftTasks, nil)
	o.outboxRepository.EXPECT().MarkAsProcessed(gomock.Any(), gomock.Any()).AnyTimes().
		DoAndReturn(func(_ context.Context, idempotencyKeys []string) error {
			o.successKeys = idempotencyKeys
			return nil
		})
}

func getOutboxData(t *testing.T) *outboxData {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockOutboxRepository := mocks.NewMockOutboxRepository(ctrl)
	mockTransactor := mocks.NewMockTransactor(ctrl)

	return &outboxData{
		outboxRepository: mockOutboxRepository,
		transactor:       mockTransactor,
		successKeys:      []string{},
	}
}
