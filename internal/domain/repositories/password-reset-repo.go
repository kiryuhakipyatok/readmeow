package repositories

import (
	"context"
	"errors"
	"readmeow/internal/domain/repositories/helpers"
	"readmeow/pkg/errs"
	"readmeow/pkg/storage"
	"time"
)

type PasswordResetRepo interface {
	AddToken(ctx context.Context, id string, token []byte, ttl time.Time) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
	CheckToken(ctx context.Context, id string, token []byte) (bool, error)
}

type passwordResetRepo struct {
	Storage *storage.Storage
}

func NewPasswordResetRepo(s *storage.Storage) PasswordResetRepo {
	return &passwordResetRepo{
		Storage: s,
	}
}

func (prr *passwordResetRepo) AddToken(ctx context.Context, id string, token []byte, ttl time.Time) error {
	op := "passwordResetRepo.AddToken"
	query := "INSERT INTO password_resets (user_id, token, expired_time) VALUES($1, $2, $3)"
	qd := helpers.NewQueryData(ctx, prr.Storage, op, query, id, token, ttl)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (prr *passwordResetRepo) DeleteExpired(ctx context.Context) error {
	op := "passwordResetRepo.DeleteExpired"
	query := "DELETE FROM password_resets WHERE expired_time <= NOW()"
	qd := helpers.NewQueryData(ctx, prr.Storage, op, query)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (prr *passwordResetRepo) Delete(ctx context.Context, id string) error {
	op := "passwordResetRepo.Delete"
	query := "DELETE FROM password_resets WHERE user_id = $1"
	qd := helpers.NewQueryData(ctx, prr.Storage, op, query, id)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (prr *passwordResetRepo) CheckToken(ctx context.Context, id string, token []byte) (bool, error) {
	op := "passwordResetRepo.CheckToken"
	var expired_time time.Time
	query := "SELECT expired_time FROM password_resets WHERE token = $1 AND user_id = $2"
	if tx, ok := storage.GetTx(ctx); ok {
		if err := tx.QueryRow(ctx, query, token, id).Scan(&expired_time); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return false, errs.ErrInvalidCode(op)
			}
			return false, errs.NewAppError(op, err)
		}
	} else {
		if err := prr.Storage.Pool.QueryRow(ctx, query, token, id).Scan(&expired_time); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return false, errs.ErrInvalidCode(op)
			}
			return false, errs.NewAppError(op, err)
		}
	}
	if time.Now().After(expired_time) {
		prr.Delete(ctx, id)
		return false, errs.NewAppError(op, errors.New("token is expired"))
	}
	return true, nil
}
