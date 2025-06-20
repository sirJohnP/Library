package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

//go:generate ../../../bin/mockgen -source=transactor.go -destination=mocks/transactor_mock.go -package=mocks

type Transactor interface {
	WithTx(ctx context.Context, function func(ctx context.Context) error) error
}

var _ Transactor = (*transactorImpl)(nil)

type transactorImpl struct {
	db *pgxpool.Pool
}

func NewTransactor(db *pgxpool.Pool) *transactorImpl {
	return &transactorImpl{
		db: db,
	}
}

func (t *transactorImpl) WithTx(ctx context.Context, function func(ctx context.Context) error) (txErr error) {
	ctxWithTx, tx, err := injectTx(ctx, t.db)
	if err != nil {
		return fmt.Errorf("cannot inject tx: %w", err)
	}

	defer func() {
		if txErr != nil {
			_ = tx.Rollback(ctxWithTx)
			return
		}

		_ = tx.Commit(ctxWithTx)
	}()

	err = function(ctxWithTx)

	if err != nil {
		return fmt.Errorf("function execution error: %w", err)
	}

	return nil
}

type txInjector struct{}

var ErrTxNotFound = errors.New("transaction not found")

func extractTX(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txInjector{}).(pgx.Tx)

	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}

func injectTx(ctx context.Context, pool *pgxpool.Pool) (context.Context, pgx.Tx, error) {
	if tx, err := extractTX(ctx); err == nil {
		return ctx, tx, nil
	}

	tx, err := pool.Begin(ctx)

	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, txInjector{}, tx), tx, err
}
