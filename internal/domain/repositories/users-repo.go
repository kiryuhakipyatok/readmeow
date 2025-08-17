package repositories

import (
	"context"
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
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	ExistanceCheck(ctx context.Context, login, email, nickname string) (bool, error)
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
	query := "INSERT INTO users (id, nickname, login, email, avatar, password, time_of_register, num_of_templates, num_of_readmes) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	if err := helpers.InsertWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, user.Id, user.Nickname, user.Login, user.Email, user.Avatar, user.Password, user.TimeOfRegister, user.NumOfTemplates, user.NumOfReadmes)); err != nil {
		return err
	}
	return nil
}

func (ur *userRepo) Get(ctx context.Context, id string) (*models.User, error) {
	op := "userRepo.Get"
	query := "SELECT id, login, email, avatar, time_of_register, num_of_templates, num_of_readmes FROM users WHERE id = $1"
	user := &models.User{}
	if err := helpers.QueryRowWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, id), user); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	op := "userRepo.GetByLogin"
	query := "SELECT id, login, email, password, avatar, time_of_register, num_of_templates, num_of_readmes FROM users WHERE login = $1"
	user := &models.User{}
	if err := helpers.QueryRowWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, login), user); err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepo) ExistanceCheck(ctx context.Context, login, email, nickname string) (bool, error) {
	op := "userRepo.ExistanceCheck"
	query := "SELECT 1 FROM users WHERE login = $1 OR email = $2 OR nickname = $3"
	var res int
	if err := helpers.QueryRowWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, login, email, nickname), &res); err != nil {
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
		return errs.ErrNotFound(op, nil)
	}
	return nil
}

func (ur *userRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "userRepo.Update"
	validFields := map[string]bool{
		"login":            true,
		"email":            true,
		"avatar":           true,
		"num_of_readmes":   true,
		"num_of_templates": true,
	}
	str := []string{}
	args := []any{}
	i := 1
	for k, v := range updates {
		if !validFields[k] {
			return errs.ErrInvalidFields(op, nil)
		}
		str = append(str, fmt.Sprintf(" %s = $%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE users SET%s WHERE id = $%d", strings.Join(str, ","), i)
	if err := helpers.DeleteOrUpdateWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, args...)); err != nil {
		return err
	}
	return nil
}

func (ur *userRepo) ChangePassword(ctx context.Context, id string, password []byte) error {
	op := "userRepo.UpdatePassword"
	query := "UPDATE users SET password=$1 WHERE id = $2"
	if err := helpers.DeleteOrUpdateWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, password, id)); err != nil {
		return err
	}
	return nil
}

func (ur *userRepo) GetPassword(ctx context.Context, id string) ([]byte, error) {
	op := "userRepo.GetPassword"
	query := "SELECT password FROM users WHERE id = $1"
	password := []byte{}
	if err := helpers.QueryRowWithTx(helpers.NewQueryData(ctx, ur.Storage, op, query, id), password); err != nil {
		return nil, err
	}
	return password, nil
}
