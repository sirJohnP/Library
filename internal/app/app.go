package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/config"
	"github.com/project/library/db"
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/controller"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/internal/usecase/outbox"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	gracefulShutdownTimeout = 3 * time.Second

	dialerTimeout   = 30 * time.Second
	dialerKeepAlive = 180 * time.Second

	transportMaxIdleConns          = 100
	transportMaxConnsPerHost       = 100
	transportIdleConnTimeout       = 90 * time.Second
	transportTLSHandshakeTimeout   = 15 * time.Second
	transportExpectContinueTimeout = 2 * time.Second
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)

	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		return
	}

	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepository(dbPool)
	outboxRepository := repository.NewOutboxRepository(dbPool)

	transactor := repository.NewTransactor(dbPool)
	runOutbox(ctx, cfg, logger, outboxRepository, transactor)

	useCases := library.New(logger, repo, repo, outboxRepository, transactor)

	ctrl := controller.New(logger, useCases, useCases)

	go runRest(ctx, cfg, logger)
	go runGrpc(cfg, logger, ctrl)

	<-ctx.Done()
	time.Sleep(gracefulShutdownTimeout)
}

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	dialer := &net.Dialer{
		Timeout:   dialerTimeout,
		KeepAlive: dialerKeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          transportMaxIdleConns,
		MaxConnsPerHost:       transportMaxConnsPerHost,
		IdleConnTimeout:       transportIdleConnTimeout,
		TLSHandshakeTimeout:   transportTLSHandshakeTimeout,
		ExpectContinueTimeout: transportExpectContinueTimeout,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}

	client := new(http.Client)
	client.Transport = transport

	globalHandler := globalOutboxHandler(client, cfg)
	outboxService := outbox.New(logger, outboxRepository, globalHandler, cfg, transactor)

	_ = outboxService.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.WaitTimeMS,
		cfg.Outbox.InProgressTTLMS,
	)
}

func globalOutboxHandler(
	client *http.Client,
	cfg *config.Config,
) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(client, cfg.Outbox.BookSendURL), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(client, cfg.Outbox.AuthorSendURL), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func bookOutboxHandler(client *http.Client, url string) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		book := entity.Book{}
		err := json.Unmarshal(data, &book)

		if err != nil {
			return fmt.Errorf("cannot deserialize data in book outbox handler: %w", err)
		}

		resp, err := client.Post(url, "text/plain", bytes.NewReader([]byte(book.ID)))

		defer func() {
			_ = resp.Body.Close()
		}()

		if err != nil {
			return fmt.Errorf("cannot send request: %w", err)
		}

		if resp != nil && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return nil
	}
}

func authorOutboxHandler(client *http.Client, url string) outbox.KindHandler {
	return func(_ context.Context, data []byte) error {
		author := entity.Author{}
		err := json.Unmarshal(data, &author)

		if err != nil {
			return fmt.Errorf("cannot deserialize data in author outbox handler: %w", err)
		}

		resp, err := client.Post(url, "text/plain", bytes.NewReader([]byte(author.ID)))

		defer func() {
			_ = resp.Body.Close()
		}()

		if err != nil {
			return fmt.Errorf("cannot send request: %w", err)
		}

		if resp != nil && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return nil
	}
}

func runRest(ctx context.Context, cfg *config.Config, logger *zap.Logger) {
	mux := grpcruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	address := "localhost:" + cfg.GRPC.Port
	err := generated.RegisterLibraryHandlerFromEndpoint(ctx, mux, address, opts)

	if err != nil {
		logger.Error("can not register grpc gateway", zap.Error(err))
		os.Exit(-1)
	}

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	logger.Info("gateway listening at port", zap.String("port", gatewayPort))

	if err = http.ListenAndServe(gatewayPort, mux); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(cfg *config.Config, logger *zap.Logger, libraryService generated.LibraryServer) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)

	if err != nil {
		logger.Error("can not open tcp socket", zap.Error(err))
		os.Exit(-1)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	generated.RegisterLibraryServer(s, libraryService)

	logger.Info("grpc server listening at port", zap.String("port", port))

	if err = s.Serve(lis); err != nil {
		logger.Error("grpc server listen error", zap.Error(err))
	}
}
