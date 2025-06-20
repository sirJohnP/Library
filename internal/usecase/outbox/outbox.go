package outbox

import (
	"context"
	"sync"
	"time"

	"github.com/project/library/config"
	"go.uber.org/zap"

	"github.com/project/library/internal/usecase/repository"
)

type GlobalHandler = func(kind repository.OutboxKind) (KindHandler, error)
type KindHandler = func(ctx context.Context, data []byte) error

type Outbox interface {
	Start(ctx context.Context, workers int, batchSize int, waitTimeMs time.Duration, inProgressTTLSeconds time.Duration) error
}

var _ Outbox = (*outboxImpl)(nil)

type outboxImpl struct {
	logger           *zap.Logger
	outboxRepository repository.OutboxRepository
	globalHandler    GlobalHandler
	cfg              *config.Config
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	globalHandler GlobalHandler,
	cfg *config.Config,
	transactor repository.Transactor,
) *outboxImpl {
	return &outboxImpl{
		logger:           logger,
		outboxRepository: outboxRepository,
		globalHandler:    globalHandler,
		cfg:              cfg,
		transactor:       transactor,
	}
}

func (o *outboxImpl) Start(
	ctx context.Context,
	workers int,
	batchSize int,
	waitTimeMs time.Duration,
	inProgressTTLSeconds time.Duration,
) error {
	wg := new(sync.WaitGroup)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go o.worker(ctx, wg, batchSize, waitTimeMs, inProgressTTLSeconds)
	}

	go func() {
		wg.Wait()
	}()

	return nil
}

func (o *outboxImpl) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	batchSize int,
	waitTimeMs time.Duration,
	inProgressTTLSeconds time.Duration,
) {
	defer wg.Done()

	for {
		time.Sleep(waitTimeMs)

		if !o.cfg.Outbox.Enabled {
			continue
		}

		err := o.transactor.WithTx(ctx, func(ctx context.Context) error {
			messages, txErr := o.outboxRepository.GetMessages(ctx, batchSize, inProgressTTLSeconds)

			if txErr != nil {
				o.logger.Error("cannot fetch messages from outbox", zap.Error(txErr))
				return txErr
			}

			o.logger.Info("entities fetched", zap.Int("size", len(messages)))

			successKeys := make([]string, 0, len(messages))

			for i := 0; i < len(messages); i++ {
				message := messages[i]
				key := messages[i].IdempotencyKey

				kindHandler, err := o.globalHandler(message.Kind)

				if err != nil {
					o.logger.Error("cannot fetch kind from outbox", zap.Error(err))
					continue
				}

				err = kindHandler(ctx, message.RawData)

				if err != nil {
					o.logger.Error("kind error", zap.Error(err))
					continue
				}

				successKeys = append(successKeys, key)
			}

			txErr = o.outboxRepository.MarkAsProcessed(ctx, successKeys)
			o.logger.Info("marked as processed", zap.Int("size", len(successKeys)))

			if txErr != nil {
				o.logger.Error("cannot mark some outbox tasks as processed", zap.Error(txErr))
				return txErr
			}

			return nil
		})

		if err != nil {
			o.logger.Error("worker error", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}
