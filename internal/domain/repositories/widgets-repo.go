package repositories

import (
	"context"
	"errors"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/pkg/storage"
)

type WidgetRepo interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error)
}

type widgetRepo struct {
	Storage *storage.Storage
}

func NewWidgetRepo(s *storage.Storage) WidgetRepo {
	return &widgetRepo{
		Storage: s,
	}
}

func (wr *widgetRepo) Get(ctx context.Context, id string) (*models.Widget, error) {
	op := "widgetRepo.Get"
	query := "SELECT * FROM widgets WHERE id = $1"
	widget := models.Widget{}
	if err := wr.Storage.Pool.QueryRow(ctx, query, id).Scan(
		&widget.Id,
		&widget.Title,
		&widget.Image,
		&widget.Description,
		&widget.Link,
		&widget.Likes,
		&widget.NumOfUsers,
	); err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return &widget, nil
}

func (wr *widgetRepo) Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error) {
	op := "widgetRepo.Fetch"
	query := "SELECT * FROM widgets ORDER BY likes DESC OFFSET $1 LIMIT $2"
	rows, err := wr.Storage.Pool.Query(ctx, query, amount*page-amount, amount)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	widgets := []models.Widget{}
	for rows.Next() {
		widget := models.Widget{}
		if err := rows.Scan(&widget.Id, &widget.Title, &widget.Image, &widget.Description, &widget.Link, &widget.NumOfUsers); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		widgets = append(widgets, widget)
	}
	return widgets, nil
}

func (wr *widgetRepo) Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error) {
	op := "widgerRepo.Filter"
	validFields := map[string]bool{
		"likes":        true,
		"num_of_users": true,
	}
	if !validFields[field] {
		return nil, fmt.Errorf("%s : %w", op, errors.New("not valid field to sort"))
	}
	if dest != "DESC" && dest != "ASC" {
		dest = "DESC"
	}
	query := fmt.Sprintf("SELECT * FROM widgets ORDER BY %s %s OFFSET $2 LIMIT $3", field, dest)
	rows, err := wr.Storage.Pool.Query(ctx, query, amount*page-amount, amount)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	widgets := []models.Widget{}
	for rows.Next() {
		widget := models.Widget{}
		if err := rows.Scan(&widget.Id, &widget.Title, &widget.Image, &widget.Description, &widget.Link, &widget.NumOfUsers); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		widgets = append(widgets, widget)
	}
	return widgets, nil
}
