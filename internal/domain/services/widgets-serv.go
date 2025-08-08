package services

import (
	"context"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
)

type WidgetServ interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error)
	Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error)
	Like(ctx context.Context, id string) error
	Dislike(ctx context.Context, id string) error
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
	ws.Logger.AddOp(op)
	ws.Logger.Log.Info("receiving widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		ws.Logger.Log.Error("failed to get widget", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ws.Logger.Log.Info("widget received successfully")
	return widget, nil
}

func (ws *widgetServ) Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error) {
	op := "widgetServ.Fetch"
	ws.Logger.AddOp(op)
	ws.Logger.Log.Info("fetching widgets")
	widgets, err := ws.WidgetRepo.Fetch(ctx, amount, page)
	if err != nil {
		ws.Logger.Log.Error("failed to fetch widgets", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ws.Logger.Log.Info("widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error) {
	op := "widgetServ.Sort"
	ws.Logger.AddOp(op)
	ws.Logger.Log.Info("fetching sorted widgets")
	widgets, err := ws.WidgetRepo.Sort(ctx, amount, page, field, dest)
	if err != nil {
		ws.Logger.Log.Error("failed to fetch sorted widgets", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ws.Logger.Log.Info("sorted widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error) {
	op := "widgetServ.Search"
	ws.Logger.AddOp(op)
	ws.Logger.Log.Info("fetching searched widgets")
	widgets, err := ws.WidgetRepo.Search(ctx, amount, page, query)
	if err != nil {
		ws.Logger.Log.Error("failed to fetch searched widgets", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ws.Logger.Log.Info("searched widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Like(ctx context.Context, id string) error {
	op := "widgetServ.Like"
	ws.Logger.AddOp(op)
	ws.Logger.Log.Info("liking widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		ws.Logger.Log.Error("failed to get widget", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	updatedLikes := widget.Likes + 1
	update := map[string]any{
		"likes": updatedLikes,
	}
	if err := ws.WidgetRepo.Update(ctx, update, widget.Id.String()); err != nil {
		ws.Logger.Log.Error("failed to update widget info", logger.Err(err))
		return err
	}
	ws.Logger.Log.Info("widget liked successfully")
	return nil
}

func (ws *widgetServ) Dislike(ctx context.Context, id string) error {
	op := "widgetServ.Dislike"
	ws.Logger.AddOp(op)
	ws.Logger.Log.Info("disliking widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		ws.Logger.Log.Error("failed to get widget", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	updatedLikes := widget.Likes - 1
	update := map[string]any{
		"likes": updatedLikes,
	}
	if err := ws.WidgetRepo.Update(ctx, update, widget.Id.String()); err != nil {
		ws.Logger.Log.Error("failed to update widget info", logger.Err(err))
		return err
	}
	ws.Logger.Log.Info("widget disliked successfully")
	return nil
}
