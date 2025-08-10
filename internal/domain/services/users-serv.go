package services

import (
	"context"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserServ interface {
	Get(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id, password string) error
	ChangePassword(ctx context.Context, id string, oldPassword, newPasswrod string) error
}

type userServ struct {
	UserRepo repositories.UserRepo
	Logger   *logger.Logger
}

func NewUserServ(ur repositories.UserRepo, l *logger.Logger) UserServ {
	return &userServ{
		UserRepo: ur,
		Logger:   l,
	}
}

func (us *userServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "userServ.Update"
	log := us.Logger.AddOp(op)
	log.Log.Info("updating user info")
	if err := us.UserRepo.Update(ctx, updates, id); err != nil {
		log.Log.Error("failed to update user info", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("user info updated successfully")
	return nil
}

func (us *userServ) Delete(ctx context.Context, id, password string) error {
	op := "userServ.Delete"
	log := us.Logger.AddOp(op)
	log.Log.Info("deleting user")
	userPassword, err := us.UserRepo.GetPassword(ctx, id)
	if err != nil {
		log.Log.Error("failed to get user password", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(userPassword, []byte(password)); err != nil {
		log.Log.Error("user and entered passwords are not equal", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	if err := us.UserRepo.Delete(ctx, id); err != nil {
		log.Log.Error("failed to delete user", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("user deleted successfully")
	return nil
}

func (us *userServ) ChangePassword(ctx context.Context, id string, oldPassword, newPasswrod string) error {
	op := "userServ.UpdatePassword"
	log := us.Logger.AddOp(op)
	log.Log.Info("changing user password")

	userPassword, err := us.UserRepo.GetPassword(ctx, id)
	if err != nil {
		log.Log.Error("failed to get user password", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(userPassword, []byte(oldPassword)); err != nil {
		log.Log.Error("old and entered passwords are not equal", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswrod), 14)
	if err != nil {
		log.Log.Error("failed to hash password", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}

	if err := us.UserRepo.ChangePassword(ctx, id, newHashedPassword); err != nil {
		log.Log.Error("failed to change user password", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("user password changed successfully")
	return nil
}

func (us *userServ) Get(ctx context.Context, id string) (*models.User, error) {
	op := "userServ.Get"
	log := us.Logger.AddOp(op)
	log.Log.Info("receiving user")
	user, err := us.UserRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get user", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("user received successfully")
	return user, nil
}
