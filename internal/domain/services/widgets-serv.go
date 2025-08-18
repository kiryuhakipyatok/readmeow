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
	Fetch(ctx context.Context, amount, page uint) ([]dto.WidgetResponse, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]dto.WidgetResponse, error)
	Search(ctx context.Context, amount, page uint, query string) ([]dto.WidgetResponse, error)
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
	log.Log.Info("receiving widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get widget", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	log.Log.Info("widget received successfully")
	return widget, nil
}

func (ws *widgetServ) Fetch(ctx context.Context, amount, page uint) ([]dto.WidgetResponse, error) {
	op := "widgetServ.Fetch"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching widgets")
	wids, err := ws.WidgetRepo.Fetch(ctx, amount, page)
	if err != nil {
		log.Log.Error("failed to fetch widgets", logger.Err(err))
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
	log.Log.Info("widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Sort(ctx context.Context, amount, page uint, field, dest string) ([]dto.WidgetResponse, error) {
	op := "widgetServ.Sort"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching sorted widgets")
	wids, err := ws.WidgetRepo.Sort(ctx, amount, page, field, dest)
	if err != nil {
		log.Log.Error("failed to fetch sorted widgets", logger.Err(err))
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
	log.Log.Info("sorted widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Search(ctx context.Context, amount, page uint, query string) ([]dto.WidgetResponse, error) {
	op := "widgetServ.Search"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching searched widgets")
	wids, err := ws.WidgetRepo.Search(ctx, amount, page, query)
	if err != nil {
		log.Log.Error("failed to fetch searched widgets", logger.Err(err))
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
	log.Log.Info("searched widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Like(ctx context.Context, id, uid string) error {
	op := "widgetServ.Like"
	log := ws.Logger.AddOp(op)
	log.Log.Info("liking widget")
	if _, err := ws.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		if err := ws.WidgetRepo.Like(c, uid, id); err != nil {
			log.Log.Error("failed to like widget", logger.Err(err))
			return nil, err
		}
		update := map[string]string{
			"likes": "+",
		}
		if err := ws.WidgetRepo.Update(c, update, id); err != nil {
			log.Log.Error("failed to update widget", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}

	log.Log.Info("widget liked successfully")
	return nil
}

func (ws *widgetServ) Dislike(ctx context.Context, id, uid string) error {
	op := "widgetServ.Dislike"
	log := ws.Logger.AddOp(op)
	log.Log.Info("disliking widget")
	if _, err := ws.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		if err := ws.WidgetRepo.Dislike(c, uid, id); err != nil {
			log.Log.Error("failed to dislike widget", logger.Err(err))
			return nil, err
		}
		update := map[string]string{
			"likes": "-",
		}
		if err := ws.WidgetRepo.Update(c, update, id); err != nil {
			log.Log.Error("failed to update widget", logger.Err(err))
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return errs.NewAppError(op, err)
	}
	log.Log.Info("widget disliked successfully")
	return nil
}

func (ws *widgetServ) FetchFavorite(ctx context.Context, id string, amount, page uint) ([]dto.WidgetResponse, error) {
	op := "widgetServ.FetchFavorite"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching favorite widgets")
	wids, err := ws.WidgetRepo.FetchFavorite(ctx, id, amount, page)
	if err != nil {
		log.Log.Error("failed to fetch favorite widgets", logger.Err(err))
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
	log.Log.Info("favorites widgets fetched successfully")
	return widgets, nil
}
