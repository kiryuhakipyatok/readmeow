package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Transactor interface {
	WithinTransaction(ctx context.Context, tFunc func(c context.Context) (any, error)) (any, error)
}

type transactor struct {
	Storage *Storage
}

func NewTransactor(s *Storage) Transactor {
	return &transactor{
		Storage: s,
	}
}

type txKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func (t *transactor) WithinTransaction(ctx context.Context, tFunc func(c context.Context) (any, error)) (any, error) {
	txCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := t.Storage.Pool.BeginTx(txCtx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	ctxTime, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := tFunc(injectTx(ctxTime, tx))
	if err != nil {
		if err := tx.Rollback(ctxTime); err != nil {
			return nil, fmt.Errorf("failed to rollback transaction: %w", err)
		}
		return nil, err
	}
	if err := tx.Commit(ctxTime); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return res, nil
}
