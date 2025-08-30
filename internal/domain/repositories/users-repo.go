package repositories

import (
	"context"
	"errors"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories/helpers"
	"readmeow/pkg/errs"
	"readmeow/pkg/storage"
	"strings"
)

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	Get(ctx context.Context, id string) (*models.User, error)
	GetByProviderId(ctx context.Context, pid, provider string) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	GetByIds(ctx context.Context, ids []string) ([]models.User, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	IdCheck(ctx context.Context, id string) (bool, error)
	GetAvatar(ctx context.Context, id string) (string, error)
	ExistanceCheck(ctx context.Context, login, email string) (bool, error)
	ChangePassword(ctx context.Context, id string, password []byte) error
	GetPassword(ctx context.Context, id string) ([]byte, error)
}

type userRepo struct {
	Storage *storage.Storage
}

func NewUserRepo(s *storage.Storage) UserRepo {
	return &userRepo{
		Storage: s,
	}
}

func (ur *userRepo) Create(ctx context.Context, user *models.User) error {
	op := "userRepo.Create"
	query := "INSERT INTO users (id, nickname, login, email, avatar, password, time_of_register, num_of_templates, num_of_readmes, provider, provider_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,$11)"
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, user.Id, user.Nickname, user.Login, user.Email, user.Avatar, user.Password, user.TimeOfRegister, user.NumOfTemplates, user.NumOfReadmes, user.Provider, user.ProviderId)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (ur *userRepo) Get(ctx context.Context, id string) (*models.User, error) {
	op := "userRepo.Get"
	query := "SELECT id, nickname, login, email, password, avatar, time_of_register, num_of_templates, num_of_readmes FROM users WHERE id = $1"
	user := &models.User{}
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, id)
	if err := qd.QueryRowWithTx(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepo) GetByIds(ctx context.Context, ids []string) ([]models.User, error) {
	op := "userRepo.GetByIds"
	query := "SELECT id, avatar, nickname FROM users WHERE id = ANY($1)"
	users := []models.User{}
	rows, err := ur.Storage.Pool.Query(ctx, query, ids)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	for rows.Next() {
		user := models.User{}
		if err := rows.Scan(
			&user.Id,
			&user.Avatar,
			&user.Nickname,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (ur *userRepo) GetAvatar(ctx context.Context, id string) (string, error) {
	op := "userRepo.GetImage"
	query := "SELECT avatar FROM users WHERE id = $1"
	var url string
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, id)
	if err := qd.QueryRowWithTx(&url); err != nil {
		return "", err
	}
	return url, nil
}

func (ur *userRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	op := "userRepo.GetByLogin"
	query := "SELECT id, nickname, login, email, password, avatar, time_of_register, num_of_templates, num_of_readmes FROM users WHERE login = $1 AND provider = 'local'"
	user := &models.User{}
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, login)
	if err := qd.QueryRowWithTx(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepo) ExistanceCheck(ctx context.Context, login, email string) (bool, error) {
	op := "userRepo.ExistanceCheck"
	query := "SELECT 1 FROM users WHERE login = $1 OR email = $2"
	var res int
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, login, email)
	if err := qd.QueryRowWithTx(&res); err != nil {
		if errors.Is(err, errs.ErrNotFoundBase) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ur *userRepo) GetByProviderId(ctx context.Context, pid, provider string) (*models.User, error) {
	op := "userRepo.GetByProviderId"
	query := "SELECT id, nickname, login, email, password, avatar, time_of_register, num_of_templates, num_of_readmes FROM users WHERE provider_id = $1 AND provider = $2"
	user := &models.User{}
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, pid, provider)
	if err := qd.QueryRowWithTx(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepo) IdCheck(ctx context.Context, id string) (bool, error) {
	op := "userRepo.IdCheck"
	query := "SELECT 1 FROM users WHERE id = $1"
	var res int
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, id)
	if err := qd.QueryRowWithTx(&res); err != nil {
		return false, err
	}
	return res == 1, nil
}

func (ur *userRepo) Delete(ctx context.Context, id string) error {
	op := "userRepo.Delete"
	query := "DELETE FROM users WHERE id = $1"
	res, err := ur.Storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (ur *userRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "userRepo.Update"
	validFields := map[string]bool{
		"nickname":         true,
		"avatar":           true,
		"num_of_readmes":   true,
		"num_of_templates": true,
	}
	str := []string{}
	args := []any{}
	i := 1
	validValuesForNumOfTemplsAndNumOfReadmes := map[string]bool{
		"+": true,
		"-": true,
	}
	for k, v := range updates {
		if !validFields[k] {
			return errs.ErrInvalidFields(op)
		}
		if k == "num_of_readmes" || k == "num_of_templates" {
			val := v.(string)
			if !validValuesForNumOfTemplsAndNumOfReadmes[val] {
				return errs.ErrInvalidFields(op)
			}
			str = append(str, fmt.Sprintf(" %s = GREATEST(%s %s 1, 0)", k, k, val))
		} else {
			str = append(str, fmt.Sprintf(" %s = $%d", k, i))
			args = append(args, v)
			i++
		}
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE users SET%s WHERE id = $%d", strings.Join(str, ","), i)
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, args...)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (ur *userRepo) ChangePassword(ctx context.Context, id string, password []byte) error {
	op := "userRepo.UpdatePassword"
	query := "UPDATE users SET password=$1 WHERE id = $2 AND provider = 'local'"
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, password, id)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (ur *userRepo) GetPassword(ctx context.Context, id string) ([]byte, error) {
	op := "userRepo.GetPassword"
	query := "SELECT password FROM users WHERE id = $1 AND provider = 'local'"
	password := []byte{}
	qd := helpers.NewQueryData(ctx, ur.Storage, op, query, id)
	if err := qd.QueryRowWithTx(&password); err != nil {
		return nil, err
	}
	return password, nil
}
