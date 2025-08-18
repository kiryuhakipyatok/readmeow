package services

import (
	"context"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/dto"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type TemplateServ interface {
	Create(ctx context.Context, oid, title, image, description string, text, links, order []string, widgets []map[string]string) error
	Update(ctx context.Context, fields map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	Fetch(ctx context.Context, amount, page uint) ([]dto.TemplateResponse, error)
	FetchFavorite(ctx context.Context, id string, amount, page uint) ([]dto.TemplateResponse, error)
	Sort(ctx context.Context, amount, page uint, dest, field string) ([]dto.TemplateResponse, error)
	Search(ctx context.Context, amount, page uint, query string) ([]dto.TemplateResponse, error)
	Like(ctx context.Context, id, uid string) error
	Dislike(ctx context.Context, id, uid string) error
}

type templateServ struct {
	TemplateRepo repositories.TemplateRepo
	UserRepo     repositories.UserRepo
	WidgetRepo   repositories.WidgetRepo
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewTemplateServ(tr repositories.TemplateRepo, ur repositories.UserRepo, wr repositories.WidgetRepo, t storage.Transactor, l *logger.Logger) TemplateServ {
	return &templateServ{
		TemplateRepo: tr,
		UserRepo:     ur,
		WidgetRepo:   wr,
		Transactor:   t,
		Logger:       l,
	}
}

func (ts *templateServ) Create(ctx context.Context, oid, title, image, description string, text, links, order []string, widgets []map[string]string) error {
	op := "templateServ.Create"
	log := ts.Logger.AddOp(op)
	log.Log.Info("creating template")
	_, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := ts.UserRepo.Get(c, oid)
		if err != nil {
			log.Log.Error("failed to get user", logger.Err(err))
			return nil, err
		}
		template := &models.Template{
			Id:             uuid.New(),
			OwnerId:        user.Id,
			Title:          title,
			Image:          image,
			Description:    description,
			Text:           text,
			Links:          links,
			Widgets:        widgets,
			Likes:          0,
			NumOfUsers:     0,
			Order:          order,
			CreateTime:     time.Now(),
			LastUpdateTime: time.Now(),
		}
		keys := make([]string, 0, len(widgets))
		for _, w := range widgets {
			for k := range w {
				keys = append(keys, k)
			}
		}
		if len(widgets) != 0 {
			widgetsData, err := ts.WidgetRepo.GetByIds(c, keys)
			if err != nil {
				log.Log.Error("failed to fetch widgets", logger.Err(err))
				return nil, err
			}

			for _, w := range widgetsData {
				update := map[string]string{
					"num_of_users": "+",
				}
				if err := ts.WidgetRepo.Update(c, update, w.Id.String()); err != nil {
					log.Log.Error("failed to update widget info", logger.Err(err))
					return nil, err
				}
			}

		}
		if err := ts.TemplateRepo.Create(c, template); err != nil {
			log.Log.Error("failed to create template", logger.Err(err))
			return nil, err
		}
		update := map[string]any{
			"num_of_templates": "+",
		}
		if err := ts.UserRepo.Update(c, update, user.Id.String()); err != nil {
			log.Log.Error("failed to update user info", logger.Err(err))
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

func (ts *templateServ) Update(ctx context.Context, fields map[string]any, id string) error {
	op := "templateServ.Update"
	log := ts.Logger.AddOp(op)
	log.Log.Info("updating template")
	if err := ts.TemplateRepo.Update(ctx, fields, id); err != nil {
		log.Log.Error("failed to update template", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("template updated successfully")
	return nil
}

func (ts *templateServ) Delete(ctx context.Context, id string) error {
	op := "templateServ.Delete"
	log := ts.Logger.AddOp(op)
	log.Log.Info("deleting template")
	if err := ts.TemplateRepo.Delete(ctx, id); err != nil {
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
			log.Log.Error("undefind template owner", logger.Err(err))
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
		}
		templates = append(templates, template)
	}
	log.Log.Info("templates fetched successfully")
	return templates, nil
}

func (ts *templateServ) Fetch(ctx context.Context, amount, page uint) ([]dto.TemplateResponse, error) {
	op := "templateServ.Fetch"
	log := ts.Logger.AddOp(op)
	log.Log.Info("fetching templates")
	templs, err := ts.TemplateRepo.Fetch(ctx, amount, page)
	if err != nil {
		log.Log.Error("failed to fetch templates", logger.Err(err))
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
			log.Log.Error("undefind template owner", logger.Err(err))
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
		}
		templates = append(templates, template)
	}
	log.Log.Info("template fetched successfully")
	return templates, nil
}

func (ts *templateServ) Sort(ctx context.Context, amount, page uint, dest, field string) ([]dto.TemplateResponse, error) {
	op := "templateServ.Sort"
	log := ts.Logger.AddOp(op)
	log.Log.Info("fetching sorted templates")
	templs, err := ts.TemplateRepo.Sort(ctx, amount, page, dest, field)
	if err != nil {
		log.Log.Error("failed to fetch sorted templates", logger.Err(err))
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
			log.Log.Error("undefind template owner", logger.Err(err))
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
		}
		templates = append(templates, template)
	}
	log.Log.Info("sorted templates fetched successfully")
	return templates, nil
}

func (ts *templateServ) Search(ctx context.Context, amount, page uint, query string) ([]dto.TemplateResponse, error) {
	op := "templateServ.Search"
	log := ts.Logger.AddOp(op)
	log.Log.Info("fetchin searched templates")
	templs, err := ts.TemplateRepo.Search(ctx, amount, page, query)
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
			log.Log.Error("undefind template owner", logger.Err(err))
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
			log.Log.Error("failed to like template", logger.Err(err))
			return nil, err
		}
		update := map[string]any{
			"likes": "+",
		}
		if err := ts.TemplateRepo.Update(c, update, id); err != nil {
			log.Log.Error("failed to update template", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
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
			log.Log.Error("failed to dislike template", logger.Err(err))
			return nil, err
		}
		update := map[string]any{
			"likes": "-",
		}
		if err := ts.TemplateRepo.Update(c, update, id); err != nil {
			log.Log.Error("failed to update template", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}
	log.Log.Info("template disliked successfully")
	return nil
}
