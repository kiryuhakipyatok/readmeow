package repositories

import (
	"context"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories/helpers"
	"readmeow/pkg/errs"
	"readmeow/pkg/storage"
	"strings"

	"github.com/google/uuid"
)

type ReadmeRepo interface {
	Create(ctx context.Context, readme *models.Readme) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Get(ctx context.Context, id string) (*models.Readme, error)
	ChangeTemplateToBase(ctx context.Context, id string) error
	FetchByUser(ctx context.Context, amount, page uint, uid string) ([]models.Readme, error)
	FetchByTemplate(ctx context.Context, tid string) ([]models.Readme, error)
}

type readmeRepo struct {
	Storage *storage.Storage
}

func NewReadmeStorage(s *storage.Storage) ReadmeRepo {
	return &readmeRepo{
		Storage: s,
	}
}

func (rr *readmeRepo) Create(ctx context.Context, readme *models.Readme) error {
	op := "readmeRepo.Create"
	query := "INSERT INTO readmes (id, owner_id, template_id, image, title, text, links, widgets, render_order, create_time, last_update_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)"
	qd := helpers.NewQueryData(ctx, rr.Storage, op, query, readme.Id, readme.OwnerId, readme.TemplateId, readme.Image, readme.Title, readme.Text, readme.Links, readme.Widgets, readme.RenderOrder, readme.CreateTime, readme.LastUpdateTime)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (rr *readmeRepo) Delete(ctx context.Context, id string) error {
	op := "readmeRepo.Delete"
	query := "DELETE FROM readmes WHERE id = $1"
	res, err := rr.Storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (rr *readmeRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "readmeRepo.Update"
	validFields := map[string]bool{
		"title":            true,
		"image":            true,
		"text":             true,
		"links":            true,
		"widgets":          true,
		"render_order":     true,
		"last_update_time": true,
	}
	str := []string{}
	args := []any{}
	i := 1

	for k, v := range updates {
		if !validFields[k] {
			return errs.ErrInvalidFields(op)
		}

		str = append(str, fmt.Sprintf(" %s = $%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE readmes SET%s WHERE id = $%d", strings.Join(str, ","), i)
	qd := helpers.NewQueryData(ctx, rr.Storage, op, query, args...)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (rr *readmeRepo) ChangeTemplateToBase(ctx context.Context, id string) error {
	op := "readmeRepo.ChangeTemplateToBase"
	baseTID := uuid.Nil.String()
	query := "UPDATE readmes SET template_id = $1 WHERE template_id = $2"
	qd := helpers.NewQueryData(ctx, rr.Storage, op, query, baseTID, id)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (rr *readmeRepo) Get(ctx context.Context, id string) (*models.Readme, error) {
	op := "readmeRepo.Get"
	query := "SELECT * FROM readmes WHERE id = $1"
	readme := &models.Readme{}
	qd := helpers.NewQueryData(ctx, rr.Storage, op, query, id)
	if err := qd.QueryRowWithTx(readme); err != nil {
		return nil, err
	}
	return readme, nil
}

func (rr *readmeRepo) FetchByUser(ctx context.Context, amount, page uint, uid string) ([]models.Readme, error) {
	op := "readmeRepo.FetchByUser"
	query := "SELECT * FROM readmes WHERE owner_id = $1 OFFSET $2 LIMIT $3"
	rows, err := rr.Storage.Pool.Query(ctx, query, uid, amount*page-amount, amount)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()
	readmes := []models.Readme{}
	for rows.Next() {
		readme := models.Readme{}
		if err := rows.Scan(
			&readme.Id,
			&readme.OwnerId,
			&readme.TemplateId,
			&readme.Image,
			&readme.Title,
			&readme.Text,
			&readme.Links,
			&readme.Widgets,
			&readme.RenderOrder,
			&readme.CreateTime,
			&readme.LastUpdateTime,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		readmes = append(readmes, readme)
	}
	return readmes, nil
}

func (rr *readmeRepo) FetchByTemplate(ctx context.Context, tid string) ([]models.Readme, error) {
	op := "readmeRepo.FetchByUser"
	query := "SELECT * FROM readmes WHERE template_id = $1"
	rows, err := rr.Storage.Pool.Query(ctx, query, tid)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()
	readmes := []models.Readme{}
	for rows.Next() {
		readme := models.Readme{}
		if err := rows.Scan(
			&readme.Id,
			&readme.OwnerId,
			&readme.TemplateId,
			&readme.Image,
			&readme.Title,
			&readme.Text,
			&readme.Links,
			&readme.Widgets,
			&readme.RenderOrder,
			&readme.CreateTime,
			&readme.LastUpdateTime,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		readmes = append(readmes, readme)
	}
	return readmes, nil
}
