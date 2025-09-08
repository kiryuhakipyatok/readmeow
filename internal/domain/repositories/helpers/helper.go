package helpers

import (
	"context"
	"errors"
	"readmeow/internal/domain/models"
	"readmeow/pkg/errs"
	"readmeow/pkg/storage"
)

type QueryData struct {
	Ctx       context.Context
	Storage   *storage.Storage
	Operation string
	Query     string
	Args      []any
}

func NewQueryData(ctx context.Context, s *storage.Storage, op string, query string, args ...any) QueryData {
	return QueryData{
		Ctx:       ctx,
		Storage:   s,
		Operation: op,
		Query:     query,
		Args:      args,
	}
}

func (qd QueryData) InsertWithTx() error {
	if tx, ok := storage.GetTx(qd.Ctx); ok {
		res, err := tx.Exec(qd.Ctx, qd.Query, qd.Args...)
		if err != nil {
			if storage.ErrorAlreadyExists(err) {
				return errs.ErrAlreadyExists(qd.Operation, err)
			}
			return errs.NewAppError(qd.Operation, err)
		}
		if res.RowsAffected() == 0 {
			return errs.ErrNotFound(qd.Operation)
		}
		return nil
	}
	res, err := qd.Storage.Pool.Exec(qd.Ctx, qd.Query, qd.Args...)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return errs.ErrAlreadyExists(qd.Operation, err)
		}
		return errs.NewAppError(qd.Operation, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(qd.Operation)
	}
	return nil
}

func (qd QueryData) DeleteOrUpdateWithTx() error {
	if tx, ok := storage.GetTx(qd.Ctx); ok {
		res, err := tx.Exec(qd.Ctx, qd.Query, qd.Args...)
		if err != nil {
			return errs.NewAppError(qd.Operation, err)
		}
		if res.RowsAffected() == 0 {
			return errs.ErrNotFound(qd.Operation)
		}
		return nil
	}
	res, err := qd.Storage.Pool.Exec(qd.Ctx, qd.Query, qd.Args...)
	if err != nil {
		return errs.NewAppError(qd.Operation, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(qd.Operation)
	}
	return nil
}

func (qd QueryData) queryRow(data ...any) error {
	if tx, ok := storage.GetTx(qd.Ctx); ok {
		if err := tx.QueryRow(qd.Ctx, qd.Query, qd.Args...).Scan(data...); err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return errs.ErrNotFound(qd.Operation)
			}
			return errs.NewAppError(qd.Operation, err)
		}
		return nil
	}
	if err := qd.Storage.Pool.QueryRow(qd.Ctx, qd.Query, qd.Args...).Scan(data...); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return errs.ErrNotFound(qd.Operation)
		}
		return errs.NewAppError(qd.Operation, err)
	}
	return nil
}

func (qd QueryData) QueryRowWithTx(entity any) error {
	switch e := entity.(type) {
	case *models.User:
		userData := []any{
			&e.Id,
			&e.Nickname,
			&e.Login,
			&e.Email,
			&e.Password,
			&e.Avatar,
			&e.TimeOfRegister,
			&e.NumOfTemplates,
			&e.NumOfReadmes,
		}
		if err := qd.queryRow(userData...); err != nil {
			return err
		}
		return nil
	case *models.Readme:
		readmeData := []any{
			&e.Id,
			&e.OwnerId,
			&e.TemplateId,
			&e.Image,
			&e.Title,
			&e.Text,
			&e.Links,
			&e.Widgets,
			&e.RenderOrder,
			&e.CreateTime,
			&e.LastUpdateTime,
		}
		if err := qd.queryRow(readmeData...); err != nil {
			return err
		}
		return nil
	case *models.Template:
		templateData := []any{
			&e.Id,
			&e.OwnerId,
			&e.Title,
			&e.Image,
			&e.Description,
			&e.Text,
			&e.Links,
			&e.Widgets,
			&e.Likes,
			&e.RenderOrder,
			&e.CreateTime,
			&e.LastUpdateTime,
			&e.NumOfUsers,
			&e.IsPublic,
		}
		if err := qd.queryRow(templateData...); err != nil {
			return err
		}
		return nil
	case *models.TemplateWithOwner:
		templateWithOwnerData := []any{
			&e.Id,
			&e.OwnerId,
			&e.Title,
			&e.Image,
			&e.Description,
			&e.Text,
			&e.Links,
			&e.Widgets,
			&e.Likes,
			&e.RenderOrder,
			&e.CreateTime,
			&e.LastUpdateTime,
			&e.NumOfUsers,
			&e.IsPublic,
			&e.OwnerNickname,
			&e.OwnerAvatar,
		}
		if err := qd.queryRow(templateWithOwnerData...); err != nil {
			return err
		}
		return nil
	case *models.Widget:
		widgetData := []any{
			&e.Id,
			&e.Title,
			&e.Image,
			&e.Description,
			&e.Type,
			&e.Tags,
			&e.Link,
			&e.Likes,
			&e.NumOfUsers,
		}
		if err := qd.queryRow(widgetData...); err != nil {
			return err
		}
		return nil
	case *models.Credentials:
		credentialsData := []any{
			&e.Nickname,
			&e.Login,
			&e.Email,
			&e.Password,
		}
		if err := qd.queryRow(credentialsData...); err != nil {
			return err
		}
		return nil
	case *int:
		if err := qd.queryRow(e); err != nil {
			return err
		}
		return nil
	case *[]byte:
		if err := qd.queryRow(e); err != nil {
			return err
		}
		return nil
	case *string:
		if err := qd.queryRow(e); err != nil {
			return err
		}
		return nil
	default:
		return errs.NewAppError(qd.Operation, errors.New("invalid entity"))
	}
}
