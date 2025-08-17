package services

import (
	"context"
	"errors"
	"readmeow/internal/delivery/dto"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type ReadmeServ interface {
	Create(ctx context.Context, tid, oid, title, image string, text, links, order []string, widgets []map[string]string) error
	Delete(ctx context.Context, id, uid string) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Get(ctx context.Context, id string) (*models.Readme, error)
	FetchByUser(ctx context.Context, amount, page uint, uid string) ([]dto.ReadmeResponse, error)
}

type readmeServ struct {
	ReadmeRepo   repositories.ReadmeRepo
	UserRepo     repositories.UserRepo
	TemplateRepo repositories.TemplateRepo
	WidgetRepo   repositories.WidgetRepo
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewReadmeServ(rr repositories.ReadmeRepo, ur repositories.UserRepo, tr repositories.TemplateRepo, wr repositories.WidgetRepo, t storage.Transactor, l *logger.Logger) ReadmeServ {
	return &readmeServ{
		ReadmeRepo:   rr,
		UserRepo:     ur,
		TemplateRepo: tr,
		WidgetRepo:   wr,
		Logger:       l,
		Transactor:   t,
	}
}

func (rs *readmeServ) Create(ctx context.Context, tid, oid, title, image string, text, links, order []string, widgets []map[string]string) error {
	op := "readmeServ.Create"
	log := rs.Logger.AddOp(op)
	log.Log.Info("creating readme")
	_, err := rs.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := rs.UserRepo.Get(c, oid)
		if err != nil {
			log.Log.Error("failed to get user", logger.Err(err))
			return nil, err
		}
		var t models.Template
		if tid != "" {
			tempalate, err := rs.TemplateRepo.Get(c, tid)
			if err != nil {
				log.Log.Error("failed to get template", logger.Err(err))
				return nil, err
			}
			updatedNumOfUsers := tempalate.NumOfUsers + 1
			update := map[string]any{
				"num_of_users": updatedNumOfUsers,
			}
			if err := rs.TemplateRepo.Update(c, update, tempalate.Id.String()); err != nil {
				log.Log.Error("failed to update template info", logger.Err(err))
				return nil, err
			}
			t = *tempalate
		} else {
			t.Id = uuid.Nil
		}

		readme := &models.Readme{
			Id:             uuid.New(),
			OwnerId:        user.Id,
			TemplateId:     t.Id,
			Title:          title,
			Image:          image,
			Text:           text,
			Links:          links,
			Widgets:        widgets,
			Order:          order,
			CreateTime:     time.Now(),
			LastUpdateTime: time.Now(),
		}
		if err := rs.ReadmeRepo.Create(c, readme); err != nil {
			log.Log.Error("failed to create readme", logger.Err(err))
			return nil, err
		}
		if len(widgets) != 0 {
			keys := make([]string, 0, len(widgets))
			for _, w := range widgets {
				for k := range w {
					keys = append(keys, k)
				}
			}
			widgetsData, err := rs.WidgetRepo.GetByIds(c, keys)
			if err != nil {
				log.Log.Error("failed to fetch widgets", logger.Err(err))
				return nil, err
			}

			for _, w := range widgetsData {
				updatedNumOfUsers := w.NumOfUsers + 1
				update := map[string]any{
					"num_of_users": updatedNumOfUsers,
				}
				if err := rs.WidgetRepo.Update(c, update, w.Id.String()); err != nil {
					log.Log.Error("failed to update widget info", logger.Err(err))
					return nil, err
				}
			}

		}
		updatedNumOfReadmes := user.NumOfReadmes + 1
		update := map[string]any{
			"num_of_readmes": updatedNumOfReadmes,
		}
		if err := rs.UserRepo.Update(c, update, user.Id.String()); err != nil {
			log.Log.Error("failed to update user info", logger.Err(err))
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		log.Log.Error("failed to create readme", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("readme created successfully")
	return nil
}

func (rs *readmeServ) Delete(ctx context.Context, id, uid string) error {
	op := "readmeServ"
	log := rs.Logger.AddOp(op)
	log.Log.Info("deleting readme")
	if _, err := rs.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := rs.UserRepo.Get(c, uid)
		if err != nil {
			log.Log.Error("failed to get user", logger.Err(err))
			return nil, err
		}
		readme, err := rs.ReadmeRepo.Get(c, id)
		if err != nil {
			log.Log.Error("failed to get readme", logger.Err(err))
			return nil, err
		}
		if readme.OwnerId != user.Id {
			log.Log.Error("failed to delete readme", logger.Err(errors.New("readme owner id and user id are not equal")))
			return nil, err
		}
		if err := rs.ReadmeRepo.Delete(c, id); err != nil {
			log.Log.Error("failed to delete readme", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}

	log.Log.Info("readme deleted successfully")
	return nil
}

func (rs *readmeServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "readmeServ.Update"
	log := rs.Logger.AddOp(op)
	log.Log.Info("updating readme")
	updates["last_update_time"] = time.Now()
	if err := rs.ReadmeRepo.Update(ctx, updates, id); err != nil {
		log.Log.Error("failed to update readme", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("readme updated successfully")
	return nil
}

func (rs *readmeServ) Get(ctx context.Context, id string) (*models.Readme, error) {
	op := "readmeServ.Get"
	log := rs.Logger.AddOp(op)
	log.Log.Info("receiving readme")
	readme, err := rs.ReadmeRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to receive readme", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	log.Log.Info("readme received successfully")
	return readme, nil
}

func (rs *readmeServ) FetchByUser(ctx context.Context, amount, page uint, uid string) ([]dto.ReadmeResponse, error) {
	op := "readmeServ.FetchByUser"
	log := rs.Logger.AddOp(op)
	log.Log.Info("receiving readmes by user")
	rdms, err := rs.ReadmeRepo.FetchByUser(ctx, amount, page, uid)
	if err != nil {
		log.Log.Error("failed to receive readmes", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	readmes := make([]dto.ReadmeResponse, 0, len(rdms))
	for _, r := range rdms {
		readme := dto.ReadmeResponse{
			Id:             r.Id.String(),
			Title:          r.Title,
			Image:          r.Image,
			LastUpdateTime: r.LastUpdateTime,
			CreateTime:     r.CreateTime,
		}
		readmes = append(readmes, readme)
	}
	log.Log.Info("readmes received successfully")
	return readmes, nil
}
