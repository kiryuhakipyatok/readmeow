package services

import (
	"context"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type ReadmeServ interface {
	Create(ctx context.Context, tid, oid, title, order string, text, links, widgets []string) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Get(ctx context.Context, id string) (*models.Readme, error)
	FetchByUser(ctx context.Context, amount, page uint, uid string) ([]models.Readme, error)
}

type readmeServ struct {
	ReadmeRepo   repositories.ReadmeRepo
	UserRepo     repositories.UserRepo
	TemplateRepo repositories.TemplateRepo
	WidgetRepo   repositories.WidgetRepo
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewReadmeServ(rr repositories.ReadmeRepo, ur repositories.UserRepo, tr repositories.TemplateRepo, wr repositories.WidgetRepo, l *logger.Logger) ReadmeServ {
	return &readmeServ{
		ReadmeRepo:   rr,
		UserRepo:     ur,
		TemplateRepo: tr,
		WidgetRepo:   wr,
		Logger:       l,
	}
}

func (rs *readmeServ) Create(ctx context.Context, oid, title, order, tid string, text, links, widgets []string) error {
	op := "readmeServ.Create"
	rs.Logger.AddOp(op)
	rs.Logger.Log.Info("creating readme")
	_, err := rs.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := rs.UserRepo.Get(ctx, oid)
		if err != nil {
			rs.Logger.Log.Error("failed to get user", logger.Err(err))
			return nil, err
		}
		readme := &models.Readme{
			Id:             uuid.New(),
			OwnerId:        user.Id,
			Title:          title,
			Text:           text,
			Links:          links,
			Widgets:        widgets,
			Order:          order,
			CreateTime:     time.Now().Unix(),
			LastUpdateTime: time.Now().Unix(),
		}
		if err := rs.ReadmeRepo.Create(ctx, readme); err != nil {
			rs.Logger.Log.Error("failed to create readme", logger.Err(err))
			return nil, err
		}
		if tid != "" {
			tempalate, err := rs.TemplateRepo.Get(ctx, tid)
			if err != nil {
				rs.Logger.Log.Error("failed to get template", logger.Err(err))
				return nil, err
			}
			updatedNumOfUsers := tempalate.NumOfUsers + 1
			update := map[string]any{
				"num_of_users": updatedNumOfUsers,
			}
			if err := rs.TemplateRepo.Update(ctx, update, tempalate.Id.String()); err != nil {
				rs.Logger.Log.Error("failed to update template info", logger.Err(err))
				return nil, err
			}
		}
		if len(widgets) != 0 {
			widgetsData, err := rs.WidgetRepo.GetByIds(ctx, widgets)
			if err != nil {
				rs.Logger.Log.Error("failed to fetch widgets", logger.Err(err))
				return nil, err
			}

			for _, w := range widgetsData {
				updatedNumOfUsers := w.NumOfUsers + 1
				update := map[string]any{
					"num_of_users": updatedNumOfUsers,
				}
				if err := rs.WidgetRepo.Update(ctx, update, w.Id.String()); err != nil {
					rs.Logger.Log.Error("failed to update widget info", logger.Err(err))
					return nil, err
				}
			}

		}
		updatedNumOfReadmes := user.NumOfReadmes + 1
		update := map[string]any{
			"num_of_readmes": updatedNumOfReadmes,
		}
		if err := rs.UserRepo.Update(ctx, update, user.Id.String()); err != nil {
			rs.Logger.Log.Error("failed to update user info", logger.Err(err))
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		rs.Logger.Log.Error("failed to create readme", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	rs.Logger.Log.Info("readme created successfully")
	return nil
}

func (rs *readmeServ) Delete(ctx context.Context, id string) error {
	op := "readmeServ"
	rs.Logger.AddOp(op)
	rs.Logger.Log.Info("deleting readme")
	if err := rs.ReadmeRepo.Delete(ctx, id); err != nil {
		rs.Logger.Log.Error("failed to delete readme", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	rs.Logger.Log.Info("readme deleted successfully")
	return nil
}

func (rs *readmeServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "readmeServ.Update"
	rs.Logger.AddOp(op)
	rs.Logger.Log.Info("updating readme")
	updates["last_update_time"] = time.Now().Unix()
	if err := rs.ReadmeRepo.Update(ctx, updates, id); err != nil {
		rs.Logger.Log.Error("failed to update readme", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	rs.Logger.Log.Info("readme updated successfully")
	return nil
}

func (rs *readmeServ) Get(ctx context.Context, id string) (*models.Readme, error) {
	op := "readmeServ.Get"
	rs.Logger.AddOp(op)
	rs.Logger.Log.Info("receiving readme")
	readme, err := rs.ReadmeRepo.Get(ctx, id)
	if err != nil {
		rs.Logger.Log.Error("failed to receive readme", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	rs.Logger.Log.Info("readme received successfully")
	return readme, nil
}

func (rs *readmeServ) FetchByUser(ctx context.Context, amount, page uint, uid string) ([]models.Readme, error) {
	op := "readmeServ.FetchByUser"
	rs.Logger.AddOp(op)
	rs.Logger.Log.Info("receiving readmes by user")
	readmes, err := rs.ReadmeRepo.FetchByUser(ctx, amount, page, uid)
	if err != nil {
		rs.Logger.Log.Error("failed to receive readmes", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	rs.Logger.Log.Info("readmes received successfully")
	return readmes, nil
}
