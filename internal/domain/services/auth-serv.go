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
	"readmeow/internal/domain/services/utils"
	em "readmeow/internal/email"
	"readmeow/pkg/cloudstorage"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthServ interface {
	Register(ctx context.Context, email, code string) error
	Login(ctx context.Context, login, password string) (*loginResponce, error)
	SendVerifyCode(ctx context.Context, email, login, nickname, password string) error
	SendNewCode(ctx context.Context, email string) error
	OAuthLogin(ctx context.Context, nickname, avatar, email, pid, provider string) (*loginResponce, error)
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
				Nickname:   credentials.Nickname,
				Login:      credentials.Login,
				Email:      credentials.Email,
				Password:   credentials.Password,
				Provider:   "local",
				ProviderId: nil,
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
	Id       string
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
		log.Log.Error("invalid credentials", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	jwtToken, ttl, err := utils.GenerateJWT(as.AuthConfig.TokenTTL, user.Id.String(), as.AuthConfig.Secret)
	if err != nil {
		log.Log.Error("failed to generate jwt token", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	loginResponce := &loginResponce{
		Id:       user.Id.String(),
		Nickname: *user.Login,
		Avatar:   user.Avatar,
		JWT:      jwtToken,
		TTL:      *ttl,
	}

	log.Log.Info("token generated successfully")
	return loginResponce, nil
}

func (as *authServ) SendVerifyCode(ctx context.Context, email, login, nickname, password string) error {
	op := "authServ.SendVerifyCode"
	log := as.Logger.AddOp(op)
	log.Log.Info("sending verify code")
	if _, err := as.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		exist, err := as.UserRepo.ExistanceCheck(c, login, email)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, errs.ErrAlreadyExists(op, nil)
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

func (as *authServ) OAuthLogin(ctx context.Context, nickname, avatar, email, pid, provider string) (*loginResponce, error) {
	op := "authServ.GoogleAuth"
	log := as.Logger.AddOp(op)
	log.Log.Info("user oauth loggining")
	res, err := as.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user := &models.User{}
		var err error
		user, err = as.UserRepo.GetByProviderId(ctx, pid, provider)
		if err != nil {
			if errors.Is(err, errs.ErrNotFoundBase) {
				user = &models.User{
					Id: uuid.New(),
					Credentials: models.Credentials{
						Nickname:   nickname,
						Login:      nil,
						Email:      email,
						Password:   nil,
						Provider:   provider,
						ProviderId: &pid,
					},
					Avatar:         avatar,
					TimeOfRegister: time.Now(),
					NumOfTemplates: 0,
					NumOfReadmes:   0,
				}
				if err := as.UserRepo.Create(c, user); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			updates := map[string]any{}
			if user.Avatar != avatar || user.Nickname != nickname {
				if user.Avatar != avatar {
					updates["avatar"] = avatar
				}
				if user.Nickname != nickname {
					updates["nickname"] = nickname
				}
				if err := as.UserRepo.Update(c, updates, user.Id.String()); err != nil {
					return nil, err
				}
			}
		}

		jwtToken, ttl, err := utils.GenerateJWT(as.AuthConfig.TokenTTL, user.Id.String(), as.AuthConfig.Secret)
		if err != nil {
			return nil, err
		}
		loginResponce := &loginResponce{
			Id:       user.Id.String(),
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			JWT:      jwtToken,
			TTL:      *ttl,
		}
		return loginResponce, nil
	})
	if err != nil {
		log.Log.Error("failed to login user with oauth", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	log.Log.Info("token generated successfully")
	return res.(*loginResponce), nil
}
