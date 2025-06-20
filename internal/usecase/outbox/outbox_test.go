package outbox

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/project/library/internal/usecase/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func correctHandler(_ context.Context, _ []byte) error {
	return nil
}

func TestOutboxWorker(t *testing.T) {
	t.Parallel()

	batchSize := 3
	sleepTime := 500 * time.Millisecond
	inProgressTTL := 1000 * time.Millisecond

	messages := make([]repository.OutboxData, 0)
	allIdempotencyKeys := make([]string, 0)
	authorIdempotencyKeys := make([]string, 0)
	for i := 0; i < 5; i++ {
		messages = append(messages, repository.OutboxData{
			IdempotencyKey: "book_" + strconv.Itoa(i),
			Kind:           repository.OutboxKindBook,
			RawData:        make([]byte, 0),
		})
		allIdempotencyKeys = append(allIdempotencyKeys, "book_"+strconv.Itoa(i))
		messages = append(messages, repository.OutboxData{
			IdempotencyKey: "author_" + strconv.Itoa(i),
			Kind:           repository.OutboxKindAuthor,
			RawData:        make([]byte, 0),
		})
		allIdempotencyKeys = append(allIdempotencyKeys, "author_"+strconv.Itoa(i))
		authorIdempotencyKeys = append(authorIdempotencyKeys, "author_"+strconv.Itoa(i))
	}

	tests := []struct {
		testName        string
		prepare         func(*outboxData) *outboxImpl
		expectedKeySent []string
	}{
		{
			testName: "worker processed successfully",
			prepare: func(data *outboxData) *outboxImpl {
				data.prepareDefaultOutboxRepository(messages)
				data.prepareDefaultTransactor()

				return data.getImpl(t, func(_ repository.OutboxKind) (KindHandler, error) {
					return correctHandler, nil
				})
			},
			expectedKeySent: allIdempotencyKeys,
		},
		{
			testName: "message fetch failed",
			prepare: func(data *outboxData) *outboxImpl {
				data.outboxRepository.EXPECT().GetMessages(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("error"))
				data.prepareDefaultTransactor()

				return data.getImpl(t, func(_ repository.OutboxKind) (KindHandler, error) {
					return correctHandler, nil
				})
			},
			expectedKeySent: []string{},
		},
		{
			testName: "some kinds are not supported",
			prepare: func(data *outboxData) *outboxImpl {
				data.prepareDefaultOutboxRepository(messages)
				data.prepareDefaultTransactor()

				return data.getImpl(t, func(kind repository.OutboxKind) (KindHandler, error) {
					switch kind {
					case repository.OutboxKindBook:
						return nil, errors.New("error")
					default:
						return correctHandler, nil
					}
				})
			},
			expectedKeySent: authorIdempotencyKeys,
		},
		{
			testName: "some kinds have bad handlers",
			prepare: func(data *outboxData) *outboxImpl {
				data.prepareDefaultOutboxRepository(messages)
				data.prepareDefaultTransactor()

				return data.getImpl(t, func(kind repository.OutboxKind) (KindHandler, error) {
					switch kind {
					case repository.OutboxKindBook:
						return func(_ context.Context, _ []byte) error {
							return errors.New("error")
						}, nil
					default:
						return correctHandler, nil
					}
				})
			},
			expectedKeySent: authorIdempotencyKeys,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())

			data := getOutboxData(t)

			wg := new(sync.WaitGroup)
			wg.Add(1)

			impl := tt.prepare(data)
			go impl.worker(ctx, wg, batchSize, sleepTime, inProgressTTL)

			time.Sleep(2 * sleepTime)
			cancel()

			wg.Wait()

			require.ElementsMatch(t, tt.expectedKeySent, data.successKeys)
		})
	}
}
