package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/dto"
	"readmeow/pkg/cloudstorage"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserServ interface {
	Get(ctx context.Context, id string) (*dto.UserResponce, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id, password string) error
	ChangePassword(ctx context.Context, id string, oldPassword, newPasswrod string) error
}

type userServ struct {
	UserRepo     repositories.UserRepo
	CloudStorage cloudstorage.CloudStorage
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewUserServ(ur repositories.UserRepo, cs cloudstorage.CloudStorage, t storage.Transactor, l *logger.Logger) UserServ {
	return &userServ{
		UserRepo:     ur,
		CloudStorage: cs,
		Transactor:   t,
		Logger:       l,
	}
}

func (us *userServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "userServ.Update"
	log := us.Logger.AddOp(op)
	log.Log.Info("updating user info")
	fileAnyH, ok := updates["avatar"]
	var (
		newPid string
		oldURL string
	)
	if ok {
		fileH := fileAnyH.(*multipart.FileHeader)
		file, err := fileH.Open()
		if err != nil {
			log.Log.Error("failed to open file of avatar", logger.Err(err))
			return errs.NewAppError(op, err)
		}
		defer file.Close()
		folder := "avatars"
		unow := time.Now().Unix()
		filename := fmt.Sprintf("%s-%d", id, unow)
		oldURL, err = us.UserRepo.GetAvatar(ctx, id)
		if err != nil {
			log.Log.Error("failed to get user avatar", logger.Err(err))
			return errs.NewAppError(op, err)
		}
		var url string
		url, newPid, err = us.CloudStorage.UploadImage(ctx, file, filename, folder)
		if err != nil {
			log.Log.Error("failed to upload avatar", logger.Err(err))
			return errs.NewAppError(op, err)
		}
		updates["avatar"] = url
	}
	if err := us.UserRepo.Update(ctx, updates, id); err != nil {
		log.Log.Error("failed to update user info", logger.Err(err))
		if cerr := us.CloudStorage.DeleteImage(ctx, newPid); cerr != nil {
			log.Log.Error("failed to delete user avatar", logger.Err(cerr))
			return errs.NewAppError(op, fmt.Errorf("%w : %w", err, cerr))
		}
		return errs.NewAppError(op, err)
	}
	if ok {
		pId := us.CloudStorage.GetPIdFromURL(oldURL)
		if pId == "" {
			log.Log.Error("failed to get pid from url")
			return errs.NewAppError(op, errors.New("failed to get pid from url"))
		}
		if err := us.CloudStorage.DeleteImage(ctx, pId); err != nil {
			log.Log.Error("failed to delete user avatar", logger.Err(err))
			return errs.NewAppError(op, err)
		}
	}
	log.Log.Info("user info updated successfully")
	return nil
}

func (us *userServ) Delete(ctx context.Context, id, password string) error {
	op := "userServ.Delete"
	log := us.Logger.AddOp(op)
	log.Log.Info("deleting user")
	if _, err := us.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		userPassword, err := us.UserRepo.GetPassword(c, id)
		if err != nil {
			log.Log.Error("failed to get user password", logger.Err(err))
			return nil, err
		}
		if err := bcrypt.CompareHashAndPassword(userPassword, []byte(password)); err != nil {
			log.Log.Error("user and entered passwords are not equal", logger.Err(err))
			return nil, err
		}
		if err := us.UserRepo.Delete(c, id); err != nil {
			log.Log.Error("failed to delete user", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}

	log.Log.Info("user deleted successfully")
	return nil
}

func (us *userServ) ChangePassword(ctx context.Context, id string, oldPassword, newPasswrod string) error {
	op := "userServ.UpdatePassword"
	log := us.Logger.AddOp(op)
	log.Log.Info("changing user password")
	if _, err := us.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		userPassword, err := us.UserRepo.GetPassword(c, id)
		if err != nil {
			log.Log.Error("failed to get user password", logger.Err(err))
			return nil, err
		}

		if err := bcrypt.CompareHashAndPassword(userPassword, []byte(oldPassword)); err != nil {
			log.Log.Error("old and entered passwords are not equal", logger.Err(err))
			return nil, err
		}

		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswrod), 12)
		if err != nil {
			log.Log.Error("failed to hash password", logger.Err(err))
			return nil, err
		}

		if err := us.UserRepo.ChangePassword(c, id, newHashedPassword); err != nil {
			log.Log.Error("failed to change user password", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}

	log.Log.Info("user password changed successfully")
	return nil
}

func (us *userServ) Get(ctx context.Context, id string) (*dto.UserResponce, error) {
	op := "userServ.Get"
	log := us.Logger.AddOp(op)
	log.Log.Info("receiving user")
	user, err := us.UserRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get user", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	userResp := &dto.UserResponce{
		Id:             user.Id.String(),
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		NumOfReadmes:   user.NumOfReadmes,
		NumOfTemplates: user.NumOfTemplates,
		TimeOfRegister: user.TimeOfRegister,
	}
	log.Log.Info("user received successfully")
	return userResp, nil
}
