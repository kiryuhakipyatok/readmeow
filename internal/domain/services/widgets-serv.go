package services

import (
	"context"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/dto"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
)

type WidgetServ interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Search(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]dto.WidgetResponse, error)
	Like(ctx context.Context, id, uid string) error
	Dislike(ctx context.Context, id, uid string) error
	FetchFavorite(ctx context.Context, id string, amount, page uint) ([]dto.WidgetResponse, error)
}

type widgetServ struct {
	WidgetRepo repositories.WidgetRepo
	UserRepo   repositories.UserRepo
	Transactor storage.Transactor
	Logger     *logger.Logger
}

func NewWidgetServ(wr repositories.WidgetRepo, ur repositories.UserRepo, t storage.Transactor, l *logger.Logger) WidgetServ {
	return &widgetServ{
		WidgetRepo: wr,
		UserRepo:   ur,
		Transactor: t,
		Logger:     l,
	}
}

func (ws *widgetServ) Get(ctx context.Context, id string) (*models.Widget, error) {
	op := "widgetServ.Get"
	log := ws.Logger.AddOp(op)
	log.Info("receiving widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		log.Error("failed to get widget", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	log.Info("widget received successfully")
	return widget, nil
}

func (ws *widgetServ) Search(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]dto.WidgetResponse, error) {
	op := "widgetServ.Search"
	log := ws.Logger.AddOp(op)
	log.Info("fetching searched widgets")
	wids, err := ws.WidgetRepo.Search(ctx, amount, page, query, filter, sort)
	if err != nil {
		log.Error("failed to fetch searched widgets", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	widgets := make([]dto.WidgetResponse, 0, len(wids))
	for _, w := range wids {
		widget := dto.WidgetResponse{
			Id:          w.Id.String(),
			Title:       w.Title,
			Description: w.Description,
			Image:       w.Image,
			Likes:       w.Likes,
			NumOfUsers:  w.NumOfUsers,
		}
		widgets = append(widgets, widget)
	}
	log.Info("searched widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Like(ctx context.Context, id, uid string) error {
	op := "widgetServ.Like"
	log := ws.Logger.AddOp(op)
	log.Info("liking widget")
	if _, err := ws.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		if err := ws.WidgetRepo.Like(c, uid, id); err != nil {
			return nil, err
		}
		update := map[string]string{
			"likes": "+",
		}
		if err := ws.WidgetRepo.Update(c, update, id); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Error("failed to like widget", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Info("widget liked successfully")
	return nil
}

func (ws *widgetServ) Dislike(ctx context.Context, id, uid string) error {
	op := "widgetServ.Dislike"
	log := ws.Logger.AddOp(op)
	log.Info("disliking widget")
	if _, err := ws.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		if err := ws.WidgetRepo.Dislike(c, uid, id); err != nil {
			return nil, err
		}
		update := map[string]string{
			"likes": "-",
		}
		if err := ws.WidgetRepo.Update(c, update, id); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Error("failed to dislike widget", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Info("widget disliked successfully")
	return nil
}

func (ws *widgetServ) FetchFavorite(ctx context.Context, id string, amount, page uint) ([]dto.WidgetResponse, error) {
	op := "widgetServ.FetchFavorite"
	log := ws.Logger.AddOp(op)
	log.Info("fetching favorite widgets")
	wids, err := ws.WidgetRepo.FetchFavorite(ctx, id, amount, page)
	if err != nil {
		log.Error("failed to fetch favorite widgets", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	widgets := make([]dto.WidgetResponse, 0, len(wids))
	for _, w := range wids {
		widget := dto.WidgetResponse{
			Id:          w.Id.String(),
			Title:       w.Title,
			Description: w.Description,
			Image:       w.Image,
			Likes:       w.Likes,
			NumOfUsers:  w.NumOfUsers,
		}
		widgets = append(widgets, widget)
	}
	log.Info("favorites widgets fetched successfully")
	return widgets, nil
}
