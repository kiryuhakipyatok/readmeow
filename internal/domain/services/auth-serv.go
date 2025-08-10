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
	Login(ctx context.Context, login, password string) (string, *time.Time, error)
	GetId(ctx context.Context, cookie string) (string, error)
}

type authServ struct {
	UserRepo   repositories.UserRepo
	AuthConfig config.AuthConfig
	Logger     *logger.Logger
}

func NewAuthServ(ur repositories.UserRepo, l *logger.Logger, cfg config.AuthConfig) AuthServ {
	return &authServ{
		UserRepo:   ur,
		Logger:     l,
		AuthConfig: cfg,
	}
}

func (as *authServ) Register(ctx context.Context, login, email, password string) error {
	op := "authServ.Register"

	log := as.Logger.AddOp(op)

	log.Log.Info("registering user")

	userPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Log.Info("failed to generate password hash", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	user := models.User{
		Id:             uuid.New(),
		Login:          login,
		Email:          email,
		Avatar:         "empty",
		Password:       userPassword,
		TimeOfRegister: time.Now(),
		NumOfTemplates: 0,
		NumOfReadmes:   0,
	}

	if err := as.UserRepo.Create(ctx, &user); err != nil {
		log.Log.Error("failed to create user", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}

	log.Log.Info("user registered successfully")

	return nil
}

func (as *authServ) Login(ctx context.Context, login, password string) (string, *time.Time, error) {
	op := "authServ.Login"
	log := as.Logger.AddOp(op)
	log.Log.Info("logining user")
	user, err := as.UserRepo.GetByLogin(ctx, login)
	if err != nil {
		log.Log.Error("failed to get user by login")
		return "", nil, fmt.Errorf("%s : %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		log.Log.Info("invalid credentials", logger.Err(err))
		return "", nil, fmt.Errorf("%s : %w", op, err)
	}
	t := time.Now().Add(time.Second * time.Duration(as.AuthConfig.TokenTTL))
	ttl := jwt.NewNumericDate(t)
	claims := jwt.MapClaims{
		"sub": user.Id.String(),
		"exp": ttl,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwt, err := token.SignedString([]byte(as.AuthConfig.Secret))
	if err != nil {
		log.Log.Error("failed to sign token", logger.Err(err))
		return "", nil, fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("tokin generated successfully")
	return jwt, &t, nil
}

func (as *authServ) GetId(ctx context.Context, cookie string) (string, error) {
	op := "authService.GetId"
	log := as.Logger.AddOp(op)
	log.Log.Info("id receiving")
	token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(as.AuthConfig.Secret), nil
	})
	if err != nil {
		log.Log.Error("failed to parse cookie", logger.Err(err))
		return "", fmt.Errorf("%s : %w", op, err)
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	id := claims.Subject
	return id, nil
}
