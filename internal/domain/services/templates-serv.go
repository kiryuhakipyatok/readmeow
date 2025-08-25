package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/dto"
	"readmeow/pkg/cloudstorage"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type TemplateServ interface {
	Create(ctx context.Context, oid, title, description string, image *multipart.FileHeader, links, order, text []string, widgets []map[string]string) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id, uid string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	FetchFavorite(ctx context.Context, id string, amount, page uint) ([]dto.TemplateResponse, error)
	Search(ctx context.Context, amount, page uint, query string, filter map[string]bool, sort map[string]string) ([]dto.TemplateResponse, error)
	Like(ctx context.Context, id, uid string) error
	Dislike(ctx context.Context, id, uid string) error
}

type templateServ struct {
	TemplateRepo repositories.TemplateRepo
	UserRepo     repositories.UserRepo
	WidgetRepo   repositories.WidgetRepo
	ReadmeRepo   repositories.ReadmeRepo
	Transactor   storage.Transactor
	CloudStorage cloudstorage.CloudStorage
	Logger       *logger.Logger
}

func NewTemplateServ(tr repositories.TemplateRepo, rr repositories.ReadmeRepo, ur repositories.UserRepo, wr repositories.WidgetRepo, t storage.Transactor, cs cloudstorage.CloudStorage, l *logger.Logger) TemplateServ {
	return &templateServ{
		TemplateRepo: tr,
		ReadmeRepo:   rr,
		UserRepo:     ur,
		WidgetRepo:   wr,
		Transactor:   t,
		CloudStorage: cs,
		Logger:       l,
	}
}

var baseTemplateId = uuid.Nil

func (ts *templateServ) Create(ctx context.Context, oid, title, description string, image *multipart.FileHeader, links, order, text []string, widgets []map[string]string) error {
	op := "templateServ.Create"
	log := ts.Logger.AddOp(op)
	log.Log.Info("creating template")
	_, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := ts.UserRepo.Get(c, oid)
		if err != nil {
			return nil, err
		}

		id := uuid.New()

		keys := make([]string, 0, len(widgets))
		for _, w := range widgets {
			for k := range w {
				keys = append(keys, k)
			}
		}
		if len(widgets) != 0 {
			widgetsData, err := ts.WidgetRepo.GetByIds(c, keys)
			if err != nil {
				return nil, err
			}

			for _, w := range widgetsData {
				update := map[string]string{
					"num_of_users": "+",
				}
				if err := ts.WidgetRepo.Update(c, update, w.Id.String()); err != nil {
					return nil, err
				}
			}

		}
		update := map[string]any{
			"num_of_templates": "+",
		}
		if err := ts.UserRepo.Update(c, update, user.Id.String()); err != nil {
			return nil, err
		}
		file, err := image.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()
		now := time.Now()
		unow := now.Unix()
		filename := fmt.Sprintf("%s-%d", id, unow)
		var (
			url string
			pid string
		)
		folder := "templates"
		url, pid, err = ts.CloudStorage.UploadImage(c, file, filename, folder)
		if err != nil {
			return nil, err
		}
		template := &models.Template{
			Id:             id,
			OwnerId:        user.Id,
			Title:          title,
			Image:          url,
			Description:    description,
			Text:           text,
			Links:          links,
			Widgets:        widgets,
			Likes:          0,
			NumOfUsers:     0,
			RenderOrder:    order,
			CreateTime:     now,
			LastUpdateTime: now,
		}
		if err := ts.TemplateRepo.Create(c, template); err != nil {
			if cerr := ts.CloudStorage.DeleteImage(c, pid); cerr != nil {
				return nil, fmt.Errorf("%w : %w", err, cerr)
			}
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		log.Log.Error("failed to create template", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("template created successfully")
	return nil
}

func (ts *templateServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "templateServ.Update"
	log := ts.Logger.AddOp(op)
	log.Log.Info("updating template")
	if _, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		fileAnyH, ok := updates["image"]
		now := time.Now()
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
			oldURL, err = ts.TemplateRepo.GetImage(c, id)
			if err != nil {
				return nil, err
			}
			folder := "templates"
			unow := now.Unix()
			filename := fmt.Sprintf("%s-%d", id, unow)
			var url string
			url, newPid, err = ts.CloudStorage.UploadImage(c, file, filename, folder)
			if err != nil {
				return nil, err
			}
			updates["image"] = url
		}
		updates["last_update_time"] = now
		if err := ts.TemplateRepo.Update(c, updates, id); err != nil {
			if cerr := ts.CloudStorage.DeleteImage(c, newPid); cerr != nil {
				return nil, fmt.Errorf("%w : %w", err, cerr)
			}
			return nil, err
		}
		if ok {
			pId, err := ts.CloudStorage.GetPIdFromURL(oldURL)
			if err != nil {
				return nil, err
			}
			if err := ts.CloudStorage.DeleteImage(c, pId); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to update tempalte", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("template updated successfully")
	return nil
}

func (ts *templateServ) Delete(ctx context.Context, id, uid string) error {
	op := "templateServ.Delete"
	log := ts.Logger.AddOp(op)
	log.Log.Info("deleting template")
	if _, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := ts.UserRepo.Get(c, uid)
		if err != nil {
			return nil, err
		}
		template, err := ts.TemplateRepo.Get(c, id)
		if err != nil {
			return nil, err
		}
		if template.OwnerId != user.Id {
			err := errors.New("user is not template owner")
			return nil, err
		}

		tupd := map[string]any{
			"num_of_users": "+",
		}

		if err := ts.TemplateRepo.Update(c, tupd, baseTemplateId.String()); err != nil {
			return nil, err
		}

		if err := ts.TemplateRepo.Delete(c, id); err != nil {
			return nil, err
		}

		pid, err := ts.CloudStorage.GetPIdFromURL(template.Image)
		if err != nil {
			return nil, err
		}
		if err := ts.CloudStorage.DeleteImage(c, pid); err != nil {
			return nil, err
		}

		return nil, nil
	}); err != nil {
		log.Log.Error("failed to delete template", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("template deleted successfully")
	return nil
}

func (ts *templateServ) Get(ctx context.Context, id string) (*models.Template, error) {
	op := "templateServ.Get"
	log := ts.Logger.AddOp(op)
	log.Log.Info("receiving template")

	template, err := ts.TemplateRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to receive template", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	owner, err := ts.UserRepo.Get(ctx, template.OwnerId.String())
	if err != nil {
		log.Log.Error("failed to receive template owner", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	template.OwnerAvatar = owner.Avatar
	template.OwnerNickname = owner.Nickname
	log.Log.Info("template received successfully")
	return template, nil
}

func (ts *templateServ) FetchFavorite(ctx context.Context, id string, amount, page uint) ([]dto.TemplateResponse, error) {
	op := "templateServ.FetchFavorite"
	log := ts.Logger.AddOp(op)
	log.Log.Info("fetching favorite templates")
	templs, err := ts.TemplateRepo.FetchFavorite(ctx, id, amount, page)
	if err != nil {
		log.Log.Error("failed to fetch favorite templates", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	templates := make([]dto.TemplateResponse, 0, len(templs))
	uids := []string{}
	for _, t := range templs {
		uids = append(uids, t.OwnerId.String())
	}
	users, err := ts.UserRepo.GetByIds(ctx, uids)
	if err != nil {
		log.Log.Error("failed to fetch owners", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	userMap := make(map[string]models.User)
	for _, user := range users {
		userMap[user.Id.String()] = user
	}

	for _, t := range templs {
		owner, ok := userMap[t.OwnerId.String()]
		if !ok {
			err := errors.New("undefind template owner")
			log.Log.Error("failed to get template owner", logger.Err(err))
			return nil, errs.NewAppError(op, err)
		}
		template := dto.TemplateResponse{
			Id:             t.Id.String(),
			Title:          t.Title,
			Image:          t.Image,
			LastUpdateTime: t.LastUpdateTime,
			NumOfUsers:     t.NumOfUsers,
			Likes:          t.Likes,
			OwnerId:        owner.Id.String(),
			OwnerAvatar:    owner.Avatar,
			OwnerNickname:  owner.Nickname,
		}
		templates = append(templates, template)
	}
	log.Log.Info("templates fetched successfully")
	return templates, nil
}

func (ts *templateServ) Search(ctx context.Context, amount, page uint, query string, filter map[string]bool, sort map[string]string) ([]dto.TemplateResponse, error) {
	op := "templateServ.Search"
	log := ts.Logger.AddOp(op)
	log.Log.Info("fetching searched templates")
	templs, err := ts.TemplateRepo.Search(ctx, amount, page, query, filter, sort)
	if err != nil {
		log.Log.Error("failed to fetch searched templates", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	templates := make([]dto.TemplateResponse, 0, len(templs))
	uids := []string{}
	for _, t := range templs {
		uids = append(uids, t.OwnerId.String())
	}
	users, err := ts.UserRepo.GetByIds(ctx, uids)
	if err != nil {
		log.Log.Error("failed to fetch owners", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	userMap := make(map[string]models.User)
	for _, user := range users {
		userMap[user.Id.String()] = user
	}

	for _, t := range templs {
		owner, ok := userMap[t.OwnerId.String()]
		if !ok {
			err := errors.New("undefind template owner")
			log.Log.Error("failed to get template owner", logger.Err(err))
			return nil, errs.NewAppError(op, err)
		}
		template := dto.TemplateResponse{
			Id:             t.Id.String(),
			Title:          t.Title,
			Image:          t.Image,
			LastUpdateTime: t.LastUpdateTime,
			NumOfUsers:     t.NumOfUsers,
			Likes:          t.Likes,
			OwnerId:        owner.Id.String(),
			OwnerAvatar:    owner.Avatar,
			OwnerNickname:  owner.Nickname,
		}
		templates = append(templates, template)
	}
	log.Log.Info("searched templates fetched successfully")
	return templates, nil
}

func (ts *templateServ) Like(ctx context.Context, id, uid string) error {
	op := "templateServ.Like"
	log := ts.Logger.AddOp(op)
	log.Log.Info("liking template")
	if _, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		if err := ts.TemplateRepo.Like(c, id, uid); err != nil {
			return nil, err
		}
		update := map[string]any{
			"likes": "+",
		}
		if err := ts.TemplateRepo.Update(c, update, id); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to like template", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("template liked successfully")
	return nil
}

func (ts *templateServ) Dislike(ctx context.Context, id, uid string) error {
	op := "templateServ.Dislike"
	log := ts.Logger.AddOp(op)
	log.Log.Info("disliking template")
	if _, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		if err := ts.TemplateRepo.Dislike(c, id, uid); err != nil {
			return nil, err
		}
		update := map[string]any{
			"likes": "-",
		}
		if err := ts.TemplateRepo.Update(c, update, id); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}
	log.Log.Info("template disliked successfully")
	return nil
}
