package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"math/rand"
	"readmeow/internal/config"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	em "readmeow/internal/email"
	"readmeow/pkg/cloudstorage"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthServ interface {
	Register(ctx context.Context, email, code string) error
	Login(ctx context.Context, login, password string) (*loginResponce, error)
	GetId(ctx context.Context, cookie string) (string, error)
	SendVerifyCode(ctx context.Context, email, login, nickname, password string) error
	SendNewCode(ctx context.Context, email string) error
}

type authServ struct {
	UserRepo         repositories.UserRepo
	VerificationRepo repositories.VerificationRepo
	Transactor       storage.Transactor
	AuthConfig       config.AuthConfig
	CloudStorage     cloudstorage.CloudStorage
	EmailSender      em.EmailSender
	Logger           *logger.Logger
}

func NewAuthServ(ur repositories.UserRepo, vr repositories.VerificationRepo, cs cloudstorage.CloudStorage, t storage.Transactor, es em.EmailSender, l *logger.Logger, cfg config.AuthConfig) AuthServ {
	return &authServ{
		UserRepo:         ur,
		VerificationRepo: vr,
		Transactor:       t,
		Logger:           l,
		EmailSender:      es,
		CloudStorage:     cs,
		AuthConfig:       cfg,
	}
}

//go:embed assets/default-ava.jpg
var defaultAvatar []byte

func (as *authServ) Register(ctx context.Context, email, code string) error {
	op := "authServ.Register"
	log := as.Logger.AddOp(op)
	log.Log.Info("registering user")
	if _, err := as.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		codeHash := sha256.Sum256([]byte(code))
		res, err := as.VerificationRepo.CodeCheck(c, email, codeHash[:])
		if err != nil {
			return nil, err
		}
		if !res {
			err := errors.New("invalid code")
			return nil, err
		}
		credentials, err := as.VerificationRepo.GetCredentials(c, email)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		unow := now.Unix()
		folder := "avatars"
		id := uuid.New()
		filename := fmt.Sprintf("%s-%d", id, unow)
		file := bytes.NewReader(defaultAvatar)
		url, pid, err := as.CloudStorage.UploadImage(ctx, file, filename, folder)
		user := models.User{
			Id: id,
			Credentials: models.Credentials{
				Nickname: credentials.Nickname,
				Login:    credentials.Login,
				Email:    credentials.Email,
				Password: credentials.Password,
			},
			Avatar:         url,
			TimeOfRegister: now,
			NumOfTemplates: 0,
			NumOfReadmes:   0,
		}
		if err := as.VerificationRepo.Delete(c, user.Email); err != nil {
			return nil, err
		}
		if err := as.UserRepo.Create(c, &user); err != nil {
			if cerr := as.CloudStorage.DeleteImage(ctx, pid); cerr != nil {
				return nil, fmt.Errorf("%w : %w", err, cerr)
			}
			return nil, err
		}

		return nil, nil
	}); err != nil {
		log.Log.Error("failed to register user", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("user registered successfully")

	return nil
}

type loginResponce struct {
	Id       uuid.UUID
	Nickname string
	Avatar   string
	JWT      string
	TTL      time.Time
}

func (as *authServ) Login(ctx context.Context, login, password string) (*loginResponce, error) {
	op := "authServ.Login"
	log := as.Logger.AddOp(op)
	log.Log.Info("logining user")

	user, err := as.UserRepo.GetByLogin(ctx, login)
	if err != nil {
		log.Log.Error("failed to get user by login", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		log.Log.Info("invalid credentials", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	now := time.Now()
	t := now.Add(as.AuthConfig.TokenTTL)
	ttl := jwt.NewNumericDate(t)
	iat := jwt.NewNumericDate(now)
	jti := uuid.New().String()
	claims := jwt.RegisteredClaims{
		Subject:   user.Id.String(),
		ExpiresAt: ttl,
		IssuedAt:  iat,
		ID:        jti,
		Issuer:    "readmeow",
		Audience:  []string{"readmeow-users"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString([]byte(as.AuthConfig.Secret))
	if err != nil {
		log.Log.Error("failed to sign token", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	loginResponce := &loginResponce{
		Id:       user.Id,
		Nickname: user.Login,
		Avatar:   user.Avatar,
		JWT:      jwtToken,
		TTL:      t,
	}

	log.Log.Info("token generated successfully")
	return loginResponce, nil
}

func (as *authServ) GetId(ctx context.Context, cookie string) (string, error) {
	op := "authService.GetId"
	log := as.Logger.AddOp(op)
	log.Log.Info("id receiving")
	token, err := jwt.ParseWithClaims(cookie, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(as.AuthConfig.Secret), nil
	})
	if err != nil {
		log.Log.Error("failed to parse cookie", logger.Err(err))
		return "", errs.NewAppError(op, err)
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	id := claims.Subject
	exist, err := as.UserRepo.IdCheck(ctx, id)
	if err != nil {
		log.Log.Error("failed to check user id", logger.Err(err))
		return "", errs.NewAppError(op, err)
	}
	if !exist {
		log.Log.Error("user not found", logger.Err(err))
		return "", errs.ErrNotFound(op)
	}
	log.Log.Info("id received successfully")
	return id, nil
}

func (as *authServ) SendVerifyCode(ctx context.Context, email, login, nickname, password string) error {
	op := "authServ.SendVerifyCode"
	log := as.Logger.AddOp(op)
	log.Log.Info("sending verify code")
	if _, err := as.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		exist, err := as.UserRepo.ExistanceCheck(c, login, email, nickname)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, errs.ErrAlreadyExists(op, err)
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		code := fmt.Sprintf("%06d", r.Intn(1000000))
		codeHash := sha256.Sum256([]byte(code))
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return nil, err
		}
		codeTTL := time.Now().Add(as.AuthConfig.CodeTTL)
		if err := as.VerificationRepo.AddCode(c, email, login, nickname, passwordHash, codeHash[:], codeTTL, as.AuthConfig.CodeAttempts); err != nil {
			return nil, err
		}
		subject := "Email Verifying"
		content, err := em.BuildEmailLetter(code)
		if err != nil {
			return nil, err
		}
		if err := as.EmailSender.SendMessage(c, subject, []byte(content), []string{email}, nil); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to send verification code", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("code sended successfully")
	return nil
}

func (as *authServ) SendNewCode(ctx context.Context, email string) error {
	op := "authServ.SendNewCode"
	log := as.Logger.AddOp(op)
	log.Log.Info("sending new verify code")
	if _, err := as.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		code := fmt.Sprintf("%06d", r.Intn(1000000))
		codeHash := sha256.Sum256([]byte(code))
		codeTTL := time.Now().Add(as.AuthConfig.CodeTTL)
		if err := as.VerificationRepo.SendNewCode(c, email, codeHash[:], codeTTL, as.AuthConfig.CodeAttempts); err != nil {
			return nil, err
		}
		subject := "Repeated Email Verifying"
		content, err := em.BuildEmailLetter(code)
		if err != nil {
			return nil, err
		}
		if err := as.EmailSender.SendMessage(c, subject, []byte(content), []string{email}, nil); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to send new verify code", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("new code sended successfully")
	return nil
}
