package repositories

import (
	"context"
	"errors"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/pkg/storage"
	"strings"
)

type ReadmeRepo interface {
	Create(ctx context.Context, readme *models.Readme) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Get(ctx context.Context, id string) (*models.Readme, error)
	FetchByUser(ctx context.Context, amount, page uint, uid string) ([]models.Readme, error)
}

type readmeRepo struct {
	Storage *storage.Storage
}

func NewReadmeStorage(s *storage.Storage) ReadmeRepo {
	return &readmeRepo{
		Storage: s,
	}
}

var (
	errReadmeNotFound      = errors.New("readme not found")
	errReadmesNotFound     = errors.New("readmes not found")
	errReadmeAlreadyExists = errors.New("readme already exists")
)

func (rr *readmeRepo) Create(ctx context.Context, readme *models.Readme) error {
	op := "readmeRepo.Create"
	query := "INSERT INTO readmes (id, owner_id, title, text, links, widgets, order, create_time, last_update_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	if _, err := rr.Storage.Pool.Exec(ctx, query, readme.Id, readme.OwnerId, readme.Text, readme.Links, readme.Widgets, readme.Order, readme.CreateTime, readme.LastUpdateTime); err != nil {
		if storage.ErrorAlreadyExists(err) {
			return fmt.Errorf("%s : %w", op, errReadmeAlreadyExists)
		}
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (rr *readmeRepo) Delete(ctx context.Context, id string) error {
	op := "readmeRepo.Delete"
	query := "DELETE FROM readmes WHERE id = $1"
	res, err := rr.Storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %w", op, errReadmeNotFound)
	}
	return nil
}

func (rr *readmeRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "readmeRepo.Update"
	validFields := map[string]bool{
		"title":            true,
		"text":             true,
		"links":            true,
		"widgets":          true,
		"order":            true,
		"last_update_time": true,
	}
	str := []string{}
	args := []any{}
	i := 0
	for k, v := range updates {
		if !validFields[k] {
			return fmt.Errorf("%s : %w", op, errors.New("not valid fields to update"))
		}
		str = append(str, fmt.Sprintf(" %s = $%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE readmes SET%s WHERE id = $%d", strings.Join(str, ","), i)
	if tx, ok := storage.GetTx(ctx); ok {
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return fmt.Errorf("%s : %w", op, errReadmeNotFound)
			}
			return fmt.Errorf("%s : %w", op, err)
		}
	} else {
		if _, err := rr.Storage.Pool.Exec(ctx, query, args...); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return fmt.Errorf("%s : %w", op, errReadmeNotFound)
			}
			return fmt.Errorf("%s : %w", op, err)
		}
	}
	return nil
}

func (rr *readmeRepo) Get(ctx context.Context, id string) (*models.Readme, error) {
	op := "readmeRepo.Get"
	query := "SELECT * FROM readmes WHERE id = $1"
	readme := models.Readme{}
	if err := rr.Storage.Pool.QueryRow(ctx, query, id).Scan(
		&readme.Id,
		&readme.OwnerId,
		&readme.Title,
		&readme.Text,
		&readme.Links,
		&readme.Widgets,
		&readme.Order,
		&readme.CreateTime,
		&readme.LastUpdateTime,
	); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, fmt.Errorf("%s : %w", op, errReadmeNotFound)
		}
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return &readme, nil
}

func (rr *readmeRepo) FetchByUser(ctx context.Context, amount, page uint, uid string) ([]models.Readme, error) {
	op := "readmeRepo.FetchByUser"
	query := "SELECT * FROM readmes OFFSET $1 LIMIT $2 WHERE owner_id = $3"
	rows, err := rr.Storage.Pool.Query(ctx, query, amount*page-amount, amount, uid)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	readmes := []models.Readme{}
	for rows.Next() {
		readme := models.Readme{}
		if err := rows.Scan(
			&readme.Id,
			&readme.OwnerId,
			&readme.Title,
			&readme.Text,
			&readme.Links,
			&readme.Widgets,
			&readme.Order,
			&readme.CreateTime,
			&readme.LastUpdateTime,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		readmes = append(readmes, readme)
	}
	if len(readmes) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errReadmesNotFound)
	}
	return readmes, nil
}
