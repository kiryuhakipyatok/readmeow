package repositories

import (
	"context"
	"errors"
	"readmeow/internal/domain/repositories/helpers"
	"readmeow/pkg/errs"
	"readmeow/pkg/storage"
	"time"
)

type VerificationRepo interface {
	AddCode(ctx context.Context, email, login, nickname string, password []byte, code []byte, ttl time.Time, attempts int) error
	CodeCheck(ctx context.Context, email string, code []byte) (bool, error)
	GetCredentials(ctx context.Context, email string) (*Credentials, error)
	SendNewCode(ctx context.Context, email string, code []byte, ttl time.Time, attempts int) error
	DeleteExpired(ctx context.Context) error
	Delete(ctx context.Context, email string) error
}

type verificationRepo struct {
	Storage *storage.Storage
}

func NewVerificationRepo(s *storage.Storage) VerificationRepo {
	return &verificationRepo{
		Storage: s,
	}
}

type Credentials struct {
	Email    string
	Login    string
	Nickname string
	Password []byte
}

func (vr *verificationRepo) AddCode(ctx context.Context, email, login, nickname string, password []byte, code []byte, ttl time.Time, attempts int) error {
	op := "verificationRepo.AddCode"
	query := "INSERT INTO verifications (email,login,nickname,password,code,expired_time, attempts) VALUES($1,$2,$3,$4,$5,$6,$7)"
	qd := helpers.NewQueryData(ctx, vr.Storage, op, query, email, login, nickname, password, code, ttl, attempts)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (vr *verificationRepo) SendNewCode(ctx context.Context, email string, code []byte, ttl time.Time, attempts int) error {
	op := "verificationRepo.SendNewCode"
	query := "UPDATE verifications SET code = $1, expired_time=$2, attempts=$3 WHERE email = $4"
	qd := helpers.NewQueryData(ctx, vr.Storage, op, query, code, ttl, attempts, email)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (vr *verificationRepo) Delete(ctx context.Context, email string) error {
	op := "verificationRepo.Delete"
	query := "DELETE FROM verifications WHERE email = $1"
	qd := helpers.NewQueryData(ctx, vr.Storage, op, query, email)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (vr *verificationRepo) DeleteExpired(ctx context.Context) error {
	op := "verificationRepo.DeleteExpired"
	query := "DELETE FROM verifications WHERE expired_time <= NOW()"
	qd := helpers.NewQueryData(ctx, vr.Storage, op, query)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (vr *verificationRepo) minusAttempts(ctx context.Context, email string) error {
	op := "verificationRepo.minusAttempts"
	query := "UPDATE verifications SET attempts = attempts - 1 WHERE email = $1"
	res, err := vr.Storage.Pool.Exec(ctx, query, email)
	if err != nil {
		if storage.CheckErr(err) {
			vr.Delete(ctx, email)
			return errs.NewAppError(op, errors.New("attempts are zero"))
		}
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (vr *verificationRepo) CodeCheck(ctx context.Context, email string, code []byte) (bool, error) {
	op := "verificationRepo.CodeCheck"
	var expired_time time.Time
	query := "SELECT expired_time FROM verifications WHERE code = $1 AND email = $2"
	if tx, ok := storage.GetTx(ctx); ok {
		if err := tx.QueryRow(ctx, query, code, email).Scan(&expired_time); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				if err := vr.minusAttempts(ctx, email); err != nil {
					return false, errs.NewAppError(op, err)
				}
				return false, nil
			}
			return false, errs.NewAppError(op, err)
		}
	} else {
		if err := vr.Storage.Pool.QueryRow(ctx, query, code, email).Scan(&expired_time); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				if err := vr.minusAttempts(ctx, email); err != nil {
					return false, errs.NewAppError(op, err)
				}
				return false, nil
			}
			return false, errs.NewAppError(op, err)
		}
	}
	if time.Now().After(expired_time) {
		vr.Delete(ctx, email)
		return false, errs.NewAppError(op, errors.New("code is expired"))
	}

	return true, nil
}

func (vr *verificationRepo) GetCredentials(ctx context.Context, email string) (*Credentials, error) {
	op := "verificationRepo.FetchCredentials"
	query := "SELECT email,login,nickname,password FROM verifications WHERE email = $1"
	creds := &Credentials{}
	qd := helpers.NewQueryData(ctx, vr.Storage, op, query, email)
	if err := qd.QueryRowWithTx(creds); err != nil {
		return nil, err
	}
	return creds, nil
}
