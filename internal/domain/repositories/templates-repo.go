package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/pkg/cache"
	"readmeow/pkg/storage"
	"strings"
	"time"
)

type TemplateRepo interface {
	Create(ctx context.Context, template *models.Template) error
	Update(ctx context.Context, fields map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Template, error)
	Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error)
}

type templateRepo struct {
	Storage *storage.Storage
	Cache   *cache.Cache
}

func NewTemplateRepo(s *storage.Storage, c *cache.Cache) TemplateRepo {
	return &templateRepo{
		Storage: s,
		Cache:   c,
	}
}

func (tr *templateRepo) Create(ctx context.Context, template *models.Template) error {
	op := "templateRepo.Create"
	query := "INSERT INTO templates (id, owner_id, title, image, text, links, widgets, order, create_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	if _, err := tr.Storage.Pool.Exec(ctx, query, template.Id, template.OwnerId, template.Title, template.Image, template.Text, template.Links, template.Widgets, template.Order, template.CreateTime); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (tr *templateRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "templdateRepo.Update"
	validFields := map[string]bool{
		"title":   true,
		"image":   true,
		"text":    true,
		"links":   true,
		"widgets": true,
		"order":   true,
	}
	str := []string{}
	args := []any{}
	i := 1
	for k, v := range updates {
		if !validFields[k] {
			return fmt.Errorf("%s : %w", op, errors.New("not valid fields to update"))
		}
		str = append(str, fmt.Sprintf(" %s = $%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE templates SET %s WHERE id = $%d", strings.Join(str, ","), i)
	if _, err := tr.Storage.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	template, err := tr.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	cache, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	ttl, err := tr.Cache.Redis.TTL(ctx, template.Id.String()).Result()
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if err := tr.Cache.Redis.Set(ctx, template.Id.String(), cache, ttl).Err(); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (tr *templateRepo) Delete(ctx context.Context, id string) error {
	op := "templateRepo.Delete"
	query := "DELETE FROM templates WHERE id = $1"
	if _, err := tr.Storage.Pool.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if err := tr.Cache.Redis.Del(ctx, id).Err(); err != nil {
		return err
	}
	return nil
}

func (tr *templateRepo) Get(ctx context.Context, id string) (*models.Template, error) {
	op := "templateRepo.Get"
	template := &models.Template{}
	cachedTemplate, err := tr.Cache.Redis.Get(ctx, id).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedTemplate), template); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		return template, nil
	}
	if err == cache.Empty {
		query := "SELECT * FROM templates WHERE id = $1"
		if err := tr.Storage.Pool.QueryRow(ctx, query, id).Scan(
			&template.Id,
			&template.OwnerId,
			&template.Title,
			&template.Image,
			&template.Text,
			&template.Links,
			&template.Widgets,
			&template.Likes,
			&template.NumOfUsers,
			&template.Order,
			&template.CreateTime,
			&template.LastUpdateTime,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
	}

	cache, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	if err := tr.Cache.Redis.Set(ctx, template.Id.String(), cache, time.Hour*24).Err(); err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return template, nil
}

func (tr *templateRepo) Fetch(ctx context.Context, amount, page uint) ([]models.Template, error) {
	op := "templateRepo.Fetch"
	query := "SELECT * FROM templates ORDER BY likes DESC OFFSET $1 LIMIT $2"
	templates := []models.Template{}
	rows, err := tr.Storage.Pool.Query(ctx, query, amount*page-amount, amount)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		template := models.Template{}
		if err := rows.Scan(
			&template.Id,
			&template.OwnerId,
			&template.Title,
			&template.Image,
			&template.Text,
			&template.Links,
			&template.Widgets,
			&template.Likes,
			&template.NumOfUsers,
			&template.Order,
			&template.CreateTime,
			&template.LastUpdateTime,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (tr *templateRepo) Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error) {
	op := "templateRepo.Sort"
	validFields := map[string]bool{
		"num_of_users": true,
		"likes":        true,
		"create_time":  true,
	}
	if !validFields[field] {
		return nil, fmt.Errorf("%s : %w", op, errors.New("not valid field to sort"))
	}
	if dest != "ASC" && dest != "DESC" {
		dest = "DESC"
	}
	templates := []models.Template{}
	query := fmt.Sprintf("SELECT * FROM templates ORDER BY %s %s OFFSET $1 LIMIT $2", field, dest)
	rows, err := tr.Storage.Pool.Query(ctx, query, field, dest, amount*page-amount, amount)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		template := models.Template{}
		if err := rows.Scan(
			&template.Id,
			&template.OwnerId,
			&template.Title,
			&template.Image,
			&template.Text,
			&template.Links,
			&template.Widgets,
			&template.Likes,
			&template.NumOfUsers,
			&template.Order,
			&template.CreateTime,
			&template.LastUpdateTime,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		templates = append(templates, template)
	}
	return templates, nil
}
