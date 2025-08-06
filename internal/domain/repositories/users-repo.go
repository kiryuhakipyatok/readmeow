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
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id string, password []byte) error
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
	query := "INSERT INTO users (id, login, email, avatar, password, time_of_register, num_of_templates) VALUES($1, $2, $3, $4, $5, $6, $7)"
	if _, err := ur.Storage.Pool.Exec(ctx, query, user.Id, user.Login, user.Email, user.Avatar, user.Password, user.TimeOfRegister, user.NumOfTemplates); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (ur *userRepo) Get(ctx context.Context, id string) (*models.User, error) {
	op := "userRepo.Get"
	query := "SELECT id, login, email, avatar, time_of_register, num_of_register FROM users WHERE id = $1"
	user := &models.User{}
	if err := ur.Storage.Pool.QueryRow(ctx, query, id).Scan(user); err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return user, nil
}

func (ur *userRepo) Delete(ctx context.Context, id string) error {
	op := "userRepo.Delete"
	query := "DELETE FROM users WHERE id = $1"
	if _, err := ur.Storage.Pool.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (ur *userRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "userRepo.Update"
	validFields := map[string]bool{
		"login":  true,
		"email":  true,
		"avatar": true,
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
	if _, err := ur.Storage.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (ur *userRepo) UpdatePassword(ctx context.Context, id string, password []byte) error {
	op := "userRepo.UpdatePassword"
	query := "UPDATE users SET password=$1 WHERE id = $2"
	if _, err := ur.Storage.Pool.Exec(ctx, query, password, id); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}
