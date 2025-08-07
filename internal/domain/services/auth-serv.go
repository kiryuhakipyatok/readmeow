package services

import (
	"context"
	"fmt"
	"readmeow/internal/config"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthServ interface {
	Register(ctx context.Context, login, email, password string) error
	Login(ctx context.Context, login, password string) (string, error)
	Profile(ctx context.Context, id string) (*models.User, error)
}

type authServ struct {
	UserRepo   repositories.UserRepo
	AuthConfig *config.AuthConfig
	Logger     *logger.Logger
}

func NewAuthServ(ur repositories.UserRepo, l *logger.Logger, cfg *config.AuthConfig) AuthServ {
	return &authServ{
		UserRepo:   ur,
		Logger:     l,
		AuthConfig: cfg,
	}
}

func (ur *authServ) Register(ctx context.Context, login, email, password string) error {
	op := "authServ.Register"

	ur.Logger.AddOp(op)

	ur.Logger.Log.Info("registering user")

	userPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		ur.Logger.Log.Info("failed to generate password hash", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	user := models.User{
		Id:             uuid.New(),
		Login:          login,
		Email:          email,
		Avatar:         "",
		Password:       userPassword,
		TimeOfRegister: time.Now().Unix(),
		NumOfTemplates: 0,
	}

	if err := ur.UserRepo.Create(ctx, &user); err != nil {
		ur.Logger.Log.Error("failed to create user", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}

	ur.Logger.Log.Info("user registered successfully")

	return nil
}

func (ur *authServ) Login(ctx context.Context, login, password string) (string, error) {
	op := "authServ.Login"
	ur.Logger.AddOp(op)
	ur.Logger.Log.Info("logining user")
	user, err := ur.UserRepo.GetByLogin(ctx, login)
	if err != nil {
		ur.Logger.Log.Error("failed to get user by login")
		return "", fmt.Errorf("%s : %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		ur.Logger.Log.Info("invalid credentials", logger.Err(err))
		return "", fmt.Errorf("%s : %w", op, err)
	}
	claims := jwt.MapClaims{
		"sub": user.Id.String(),
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(ur.AuthConfig.TokenTTL))),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	jwt, err := token.SignedString([]byte(ur.AuthConfig.Secret))
	if err != nil {
		ur.Logger.Log.Error("failed to sign token", logger.Err(err))
		return "", fmt.Errorf("%s : %w", op, err)
	}
	ur.Logger.Log.Info("tokin generated successfully")
	return jwt, nil
}

func (ur *authServ) Profile(ctx context.Context, id string) (*models.User, error) {
	op := "authService.Profile"
	ur.Logger.AddOp(op)
	ur.Logger.Log.Info("profile receiving")
	user, err := ur.UserRepo.Get(ctx, id)
	if err != nil {
		ur.Logger.Log.Error("failed to get user", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return user, nil
}
