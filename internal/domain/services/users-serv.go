package services

import (
	"context"
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
	Get(ctx context.Context, id string) (*dto.UserResponse, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id, password string) error
	ChangePassword(ctx context.Context, id string, oldPassword, newPasswrod string) error
}

type userServ struct {
	UserRepo     repositories.UserRepo
	TemplateRepo repositories.TemplateRepo
	CloudStorage cloudstorage.CloudStorage
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewUserServ(ur repositories.UserRepo, tr repositories.TemplateRepo, cs cloudstorage.CloudStorage, t storage.Transactor, l *logger.Logger) UserServ {
	return &userServ{
		UserRepo:     ur,
		TemplateRepo: tr,
		CloudStorage: cs,
		Transactor:   t,
		Logger:       l,
	}
}

func (us *userServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "userServ.Update"
	log := us.Logger.AddOp(op)
	log.Log.Info("updating user info")
	if _, err := us.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		fileAnyH, ok := updates["avatar"]
		var (
			newPid string
			oldURL string
		)
		if ok {
			fileH := fileAnyH.(*multipart.FileHeader)
			file, err := fileH.Open()
			if err != nil {
				return nil, err
			}
			defer file.Close()
			folder := "avatars"
			unow := time.Now().Unix()
			filename := fmt.Sprintf("%s-%d", id, unow)
			oldURL, err = us.UserRepo.GetAvatar(c, id)
			if err != nil {
				return nil, err
			}
			var url string
			url, newPid, err = us.CloudStorage.UploadImage(c, file, filename, folder)
			if err != nil {
				return nil, err
			}
			updates["avatar"] = url
		}
		if err := us.UserRepo.Update(c, updates, id); err != nil {
			if cerr := us.CloudStorage.DeleteImage(c, newPid); cerr != nil {

				return nil, fmt.Errorf("%w : %w", err, cerr)
			}
			return nil, err
		}
		if ok {
			pId, err := us.CloudStorage.GetPIdFromURL(oldURL)
			if err != nil {
				return nil, err
			}
			if err := us.CloudStorage.DeleteImage(c, pId); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to update user", logger.Err(err))
		return errs.NewAppError(op, err)
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
			return nil, err
		}
		if err := bcrypt.CompareHashAndPassword(userPassword, []byte(password)); err != nil {
			return nil, err
		}
		if err := us.UserRepo.Delete(c, id); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to delete user", logger.Err(err))
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
			return nil, err
		}

		if err := bcrypt.CompareHashAndPassword(userPassword, []byte(oldPassword)); err != nil {
			return nil, err
		}

		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswrod), 12)
		if err != nil {
			return nil, err
		}

		if err := us.UserRepo.ChangePassword(c, id, newHashedPassword); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to change user password", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("user password changed successfully")
	return nil
}

func (us *userServ) Get(ctx context.Context, id string) (*dto.UserResponse, error) {
	op := "userServ.Get"
	log := us.Logger.AddOp(op)
	log.Log.Info("receiving user")
	user, err := us.UserRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get user", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	templates, err := us.TemplateRepo.FetchByUser(ctx, id)
	if err != nil {
		log.Log.Error("failed to fetch user templates", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	templateInfo := make([]dto.TemplateInfo, 0, len(templates))
	for _, t := range templates {
		temlInf := dto.TemplateInfo{
			Id:             t.Id.String(),
			Title:          t.Title,
			Description:    t.Description,
			Image:          t.Image,
			LastUpdateTime: t.LastUpdateTime,
			NumOfUsers:     t.NumOfUsers,
			Likes:          t.Likes,
		}
		templateInfo = append(templateInfo, temlInf)
	}
	userResp := &dto.UserResponse{
		Id:             user.Id.String(),
		Nickname:       user.Nickname,
		Email:          user.Email,
		Avatar:         user.Avatar,
		NumOfReadmes:   user.NumOfReadmes,
		NumOfTemplates: user.NumOfTemplates,
		TimeOfRegister: user.TimeOfRegister,
		Templates:      templateInfo,
	}
	log.Log.Info("user received successfully")
	return userResp, nil
}
