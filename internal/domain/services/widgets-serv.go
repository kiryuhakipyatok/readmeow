package services

import (
	"context"
	"errors"
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
	log := ws.Logger.AddOp(op)
	log.Log.Info("receiving widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get widget", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("widget received successfully")
	return widget, nil
}

func (ws *widgetServ) Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error) {
	op := "widgetServ.Fetch"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching widgets")
	widgets, err := ws.WidgetRepo.Fetch(ctx, amount, page)
	if err != nil {
		log.Log.Error("failed to fetch widgets", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error) {
	op := "widgetServ.Sort"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching sorted widgets")
	widgets, err := ws.WidgetRepo.Sort(ctx, amount, page, field, dest)
	if err != nil {
		log.Log.Error("failed to fetch sorted widgets", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("sorted widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error) {
	op := "widgetServ.Search"
	log := ws.Logger.AddOp(op)
	log.Log.Info("fetching searched widgets")
	widgets, err := ws.WidgetRepo.Search(ctx, amount, page, query)
	if err != nil {
		log.Log.Error("failed to fetch searched widgets", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	log.Log.Info("searched widgets fetched successfully")
	return widgets, nil
}

func (ws *widgetServ) Like(ctx context.Context, id string) error {
	op := "widgetServ.Like"
	log := ws.Logger.AddOp(op)
	log.Log.Info("liking widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get widget", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	fmt.Println(widget.Likes)
	updatedLikes := widget.Likes + 1
	fmt.Println(updatedLikes)
	update := map[string]any{
		"likes": updatedLikes,
	}
	if err := ws.WidgetRepo.Update(ctx, update, widget.Id.String()); err != nil {
		log.Log.Error("failed to update widget info", logger.Err(err))
		return err
	}
	log.Log.Info("widget liked successfully")
	return nil
}

func (ws *widgetServ) Dislike(ctx context.Context, id string) error {
	op := "widgetServ.Dislike"
	log := ws.Logger.AddOp(op)
	log.Log.Info("disliking widget")
	widget, err := ws.WidgetRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to get widget", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	if widget.Likes == 0 {
		log.Log.Error("widget likes are equal zero")
		return fmt.Errorf("%s : %w", op, errors.New("widget likes are equal zero"))
	}

	updatedLikes := widget.Likes - 1

	fmt.Println(updatedLikes)
	update := map[string]any{
		"likes": updatedLikes,
	}

	if err := ws.WidgetRepo.Update(ctx, update, widget.Id.String()); err != nil {
		log.Log.Error("failed to update widget info", logger.Err(err))
		return err
	}

	log.Log.Info("widget disliked successfully")
	return nil
}
