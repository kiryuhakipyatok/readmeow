package services

import (
	"context"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
)

type WidgetServ interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error)
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
