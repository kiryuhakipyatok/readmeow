package repositories

import (
	"context"
	"errors"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/pkg/storage"
	"strings"
)

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	Get(ctx context.Context, id string) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	ChangePassword(ctx context.Context, id string, password []byte) error
}

type userRepo struct {
	Storage *storage.Storage
}

func NewUserRepo(s *storage.Storage) UserRepo {
	return &userRepo{
		Storage: s,
	}
}

var (
	errUserNotFound      = errors.New("user not found")
	errUserAlreadyExists = errors.New("user already exists")
)

func (ur *userRepo) Create(ctx context.Context, user *models.User) error {
	op := "userRepo.Create"
	query := "INSERT INTO users (id, login, email, avatar, password, time_of_register, num_of_templates, num_of_readmes) VALUES($1, $2, $3, $4, $5, $6, $7, $8)"
	if _, err := ur.Storage.Pool.Exec(ctx, query, user.Id, user.Login, user.Email, user.Avatar, user.Password, user.TimeOfRegister, user.NumOfTemplates, user.NumOfReadmes); err != nil {
		if storage.ErrorAlreadyExists(err) {
			return fmt.Errorf("%s : %w", op, errUserAlreadyExists)
		}
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (ur *userRepo) Get(ctx context.Context, id string) (*models.User, error) {
	op := "userRepo.Get"
	query := "SELECT id, login, email, avatar, time_of_register, num_of_templates FROM users WHERE id = $1"
	user := models.User{}
	if tx, ok := storage.GetTx(ctx); ok {
		if err := tx.QueryRow(ctx, query, id).Scan(
			&user.Id,
			&user.Login,
			&user.Email,
			&user.Avatar,
			&user.TimeOfRegister,
			&user.NumOfTemplates,
			&user.NumOfReadmes,
		); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return nil, fmt.Errorf("%s : %w", op, errUserNotFound)
			}
			return nil, fmt.Errorf("%s : %w", op, err)
		}
	} else {
		if err := ur.Storage.Pool.QueryRow(ctx, query, id).Scan(
			&user.Id,
			&user.Login,
			&user.Email,
			&user.Avatar,
			&user.TimeOfRegister,
			&user.NumOfTemplates,
			&user.NumOfReadmes,
		); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return nil, fmt.Errorf("%s : %w", op, errUserNotFound)
			}
			return nil, fmt.Errorf("%s : %w", op, err)
		}
	}
	return &user, nil
}

func (ur *userRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	op := "userRepo.GetByLogin"
	query := "SELECT id, login, email, password, avatar, time_of_register, num_of_templates, num_of_readmes FROM users WHERE login = $1"
	user := models.User{}
	if err := ur.Storage.Pool.QueryRow(ctx, query, login).Scan(
		&user.Id,
		&user.Login,
		&user.Email,
		&user.Password,
		&user.Avatar,
		&user.TimeOfRegister,
		&user.NumOfTemplates,
		&user.NumOfReadmes,
	); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, fmt.Errorf("%s : %w", op, errUserNotFound)
		}
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return &user, nil
}

func (ur *userRepo) Delete(ctx context.Context, id string) error {
	op := "userRepo.Delete"
	query := "DELETE FROM users WHERE id = $1"
	res, err := ur.Storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %w", op, errUserNotFound)
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
			return fmt.Errorf("%s : %w", op, errors.New("not valid fields to update"))
		}
		str = append(str, fmt.Sprintf(" %s = $%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE users SET%s WHERE id = $%d", strings.Join(str, ","), i)
	if tx, ok := storage.GetTx(ctx); ok {
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return fmt.Errorf("%s : %w", op, errUserNotFound)
			}
			return fmt.Errorf("%s : %w", op, err)
		}
	} else {
		if _, err := ur.Storage.Pool.Exec(ctx, query, args...); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return fmt.Errorf("%s : %w", op, errUserNotFound)
			}
			return fmt.Errorf("%s : %w", op, err)
		}
	}
	return nil
}

func (ur *userRepo) ChangePassword(ctx context.Context, id string, password []byte) error {
	op := "userRepo.UpdatePassword"
	query := "UPDATE users SET password=$1 WHERE id = $2"
	if _, err := ur.Storage.Pool.Exec(ctx, query, password, id); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return fmt.Errorf("%s : %w", op, errUserNotFound)
		}
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}
