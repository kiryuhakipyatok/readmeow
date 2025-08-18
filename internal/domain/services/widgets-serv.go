package services

import (
	"context"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/dto"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
)

type WidgetServ interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Fetch(ctx context.Context, amount, page uint) ([]dto.WidgetResponse, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]dto.WidgetResponse, error)
	Search(ctx context.Context, amount, page uint, query string) ([]dto.WidgetResponse, error)
	Like(ctx context.Context, id, uid string) error
	Dislike(ctx context.Context, id, uid string) error
	FetchFavorite(ctx context.Context, id string) ([]dto.WidgetResponse, error)
}

type widgetServ struct {
	WidgetRepo repositories.WidgetRepo
	UserRepo   repositories.UserRepo
	Logger     *logger.Logger
}

func NewWidgetServ(wr repositories.WidgetRepo, ur repositories.UserRepo, l *logger.Logger) WidgetServ {
	return &widgetServ{
		WidgetRepo: wr,
		UserRepo:   ur,
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
	if err := ws.WidgetRepo.Like(ctx, uid, id); err != nil {
		log.Log.Error("failed to like widget", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("widget liked successfully")
	return nil
}

func (ws *widgetServ) Dislike(ctx context.Context, id, uid string) error {
	op := "widgetServ.Dislike"
	log := ws.Logger.AddOp(op)
	log.Log.Info("disliking widget")
	if err := ws.WidgetRepo.Dislike(ctx, uid, id); err != nil {
		log.Log.Error("failed to dislike widget", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("widget disliked successfully")
	return nil
}

func (ws *widgetServ) FetchFavorite(ctx context.Context, id string) ([]dto.WidgetResponse, error) {
	op := "widgetServ.FetchFavorite"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching favorite widgets")
	wids, err := ws.WidgetRepo.FetchFavorite(ctx, id)
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
