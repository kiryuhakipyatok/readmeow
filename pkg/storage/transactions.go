package storage

import (
	"context"
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
	tx, err := t.Storage.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			rbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tx.Rollback(rbCtx)
		}
	}()
	res, err := tFunc(injectTx(ctx, tx))
	if err != nil {
		return nil, err
	}
	cmCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tx.Commit(cmCtx); err != nil {
		return nil, err
	}
	return res, nil
}
