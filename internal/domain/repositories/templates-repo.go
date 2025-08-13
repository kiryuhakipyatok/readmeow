package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"readmeow/internal/config"
	"readmeow/internal/domain/models"
	"readmeow/pkg/cache"
	"readmeow/pkg/search"
	"readmeow/pkg/storage"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9/esutil"
	s "github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/google/uuid"
)

type TemplateRepo interface {
	Create(ctx context.Context, template *models.Template) error
	Update(ctx context.Context, fields map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Template, error)
	Like(ctx context.Context, id, uid string) error
	Dislike(ctx context.Context, id, uid string) error
	FetchFavorite(ctx context.Context, id string) ([]string, error)
	Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error)
	GetByIds(ctx context.Context, ids []string) ([]models.Template, error)
	Search(ctx context.Context, amount, page uint, query string) ([]models.Template, error)
	MustBulk(cfg config.SearchConfig)
}

type templateRepo struct {
	Storage      *storage.Storage
	Cache        *cache.Cache
	SearchClient *search.SearchClient
}

func NewTemplateRepo(s *storage.Storage, c *cache.Cache, sc *search.SearchClient) TemplateRepo {
	return &templateRepo{
		Storage:      s,
		Cache:        c,
		SearchClient: sc,
	}
}

var (
	errTemplateNotFound      = errors.New("template not found")
	errTemplatesNotFound     = errors.New("templates not found")
	errTemplateAlreadyExists = errors.New("template already exists")
)

func (tr *templateRepo) Create(ctx context.Context, template *models.Template) error {
	op := "templateRepo.Create"
	query := "INSERT INTO templates (id, owner_id, title, image,description, text, links, widgets,num_of_users, render_order, create_time, last_update_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)"
	if _, err := tr.Storage.Pool.Exec(ctx, query, template.Id, template.OwnerId, template.Title, template.Image, template.Description, template.Text, template.Links, template.Widgets, template.NumOfUsers, template.Order, template.CreateTime, template.LastUpdateTime); err != nil {
		if storage.ErrorAlreadyExists(err) {
			return fmt.Errorf("%s : %w", op, errTemplateAlreadyExists)
		}
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (tr *templateRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "templdateRepo.Update"
	validFields := map[string]bool{
		"title":            true,
		"text":             true,
		"links":            true,
		"widgets":          true,
		"order":            true,
		"num_of_users":     true,
		"last_update_time": true,
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
	if tx, ok := storage.GetTx(ctx); ok {
		res, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("%s : %w", op, errTemplateNotFound)
		}
	} else {
		res, err := tr.Storage.Pool.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("%s : %w", op, errTemplateNotFound)
		}
	}
	if err := tr.Cache.Redis.Del(ctx, id).Err(); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}

func (tr *templateRepo) Like(ctx context.Context, id, uid string) error {
	op := "templateRepo.Like"
	query := "INSERT INTO favorite_templates (template_id, user_id) VALUES ($1,$2)"
	if tx, ok := storage.GetTx(ctx); ok {
		res, err := tx.Exec(ctx, query, id, uid)
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("%s : %w", op, errTemplateNotFound)
		}
		return nil
	}
	res, err := tr.Storage.Pool.Exec(ctx, query, id, uid)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %w", op, errTemplateNotFound)
	}
	return nil
}

func (tr *templateRepo) Dislike(ctx context.Context, id, uid string) error {
	op := "templateRepo.Dislike"
	query := "DELETE FROM favorite_templates WHERE (template_id, user_id) = ($1,$2)"
	if tx, ok := storage.GetTx(ctx); ok {
		res, err := tx.Exec(ctx, query, id, uid)
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("%s : %w", op, errTemplateNotFound)
		}
		return nil
	}
	res, err := tr.Storage.Pool.Exec(ctx, query, id, uid)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %w", op, errTemplateNotFound)
	}
	return nil
}

func (tr *templateRepo) Delete(ctx context.Context, id string) error {
	op := "templateRepo.Delete"
	query := "DELETE FROM templates WHERE id = $1"
	res, err := tr.Storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%s : %w", op, errTemplateNotFound)
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
	if err == cache.EMPTY {
		query := "SELECT t.*, COUNT(ft.tempalate_id) FROM templates t LEFT JOIN favorite_templates ft ON ft.template_id=t.id WHERE id = $1 GROUP BY t.id"
		if tx, ok := storage.GetTx(ctx); ok {
			if err := tx.QueryRow(ctx, query, id).Scan(
				&template.Id,
				&template.OwnerId,
				&template.Title,
				&template.Image,
				&template.Description,
				&template.Text,
				&template.Links,
				&template.Widgets,
				&template.NumOfUsers,
				&template.Order,
				&template.CreateTime,
				&template.LastUpdateTime,
				&template.Likes,
			); err != nil {
				if errors.Is(err, storage.ErrNotFound()) {
					return nil, fmt.Errorf("%s : %w", op, errTemplateNotFound)
				}
				return nil, fmt.Errorf("%s : %w", op, err)
			}
		} else {
			if err := tr.Storage.Pool.QueryRow(ctx, query, id).Scan(
				&template.Id,
				&template.OwnerId,
				&template.Title,
				&template.Image,
				&template.Description,
				&template.Text,
				&template.Links,
				&template.Widgets,
				&template.NumOfUsers,
				&template.Order,
				&template.CreateTime,
				&template.LastUpdateTime,
				&template.Likes,
			); err != nil {
				if errors.Is(err, storage.ErrNotFound()) {
					return nil, fmt.Errorf("%s : %w", op, errTemplateNotFound)
				}
				return nil, fmt.Errorf("%s : %w", op, err)
			}
		}

	}
	if (template.NumOfUsers >= 20) || template.OwnerId == uuid.Nil {
		cache, err := json.Marshal(template)
		if err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		if err := tr.Cache.Redis.Set(ctx, template.Id.String(), cache, time.Hour*24).Err(); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
	}
	return template, nil
}

func (tr *templateRepo) FetchFavorite(ctx context.Context, id string) ([]string, error) {
	op := "templateRepo.FetchFavorite"
	query := "SELECT template_id FROM favorite_templates WHERE user_id=$1"
	tids := []string{}
	rows, err := tr.Storage.Pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errTemplatesNotFound)
	}
	return tids, nil
}

func (tr *templateRepo) Fetch(ctx context.Context, amount, page uint) ([]models.Template, error) {
	op := "templateRepo.Fetch"
	query := "SELECT t.*, COUNT(ft.tempalate_id) FROM templates t LEFT JOIN favorite_templates ft ON ft.template_id=t.id WHERE id = $1 GROUP BY t.id ORDER BY likes DESC OFFSET $1 LIMIT $2"
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
			&template.Description,
			&template.Text,
			&template.Links,
			&template.Widgets,
			&template.NumOfUsers,
			&template.Order,
			&template.CreateTime,
			&template.LastUpdateTime,
			&template.Likes,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		templates = append(templates, template)
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errTemplatesNotFound)
	}
	return templates, nil
}

func (tr *templateRepo) Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error) {
	op := "templateRepo.Sort"
	validFields := map[string]bool{
		"num_of_users": true,
		"create_time":  true,
	}
	if !validFields[field] {
		return nil, fmt.Errorf("%s : %w", op, errors.New("not valid field to sort"))
	}
	if dest != "ASC" && dest != "DESC" {
		dest = "DESC"
	}
	templates := []models.Template{}
	query := fmt.Sprintf("SELECT t.*, COUNT(ft.tempalate_id) FROM templates t LEFT JOIN favorite_templates ft ON ft.template_id=t.id WHERE id = $1 GROUP BY t.id ORDER BY %s %s OFFSET $1 LIMIT $2", field, dest)
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
			&template.Description,
			&template.Text,
			&template.Links,
			&template.Widgets,
			&template.NumOfUsers,
			&template.Order,
			&template.CreateTime,
			&template.LastUpdateTime,
			&template.Likes,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		templates = append(templates, template)
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errTemplatesNotFound)
	}
	return templates, nil
}

func (tr *templateRepo) Search(ctx context.Context, amount, page uint, query string) ([]models.Template, error) {
	op := "templateRepo.Search"
	if strings.TrimSpace(query) == "" {
		return tr.getAll(ctx)
	}
	mainQuery := types.Query{
		MultiMatch: &types.MultiMatchQuery{
			Query:     query,
			Fields:    []string{"Title^2", "Description"},
			Fuzziness: "AUTO",
		},
	}
	res, err := tr.SearchClient.Client.Search().Index("templates").From(int(amount*page - amount)).Size(int(amount)).Request(&s.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: []types.Query{mainQuery},
			},
		},
		Source_: &types.SourceFilter{Includes: []string{"id"}},
	}).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ids := []string{}
	for _, hit := range res.Hits.Hits {
		if hit.Id_ != nil {
			ids = append(ids, *hit.Id_)
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errTemplatesNotFound)
	}
	templates, err := tr.GetByIds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return templates, nil
}

func (tr *templateRepo) GetByIds(ctx context.Context, ids []string) ([]models.Template, error) {
	fmt.Println(ids)
	op := "templateRepo.SearchPreparing.GetByIds"
	query := "SELECT t.*, COUNT(ft.template_id) FROM templates t LEFT JOIN favorite_templates ft ON ft.template_id=t.id WHERE ft.template_id = ANY($1) GROUP BY t.id"
	templates := make([]models.Template, 0, len(ids))
	rows, err := tr.Storage.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	byId := map[string]models.Template{}
	for rows.Next() {
		template := models.Template{}
		if err := rows.Scan(
			&template.Id,
			&template.OwnerId,
			&template.Title,
			&template.Image,
			&template.Description,
			&template.Text,
			&template.Links,
			&template.Order,
			&template.CreateTime,
			&template.LastUpdateTime,
			&template.NumOfUsers,
			&template.Widgets,
			&template.Likes,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		byId[template.Id.String()] = template
	}
	for _, id := range ids {
		if t, ok := byId[id]; ok {
			templates = append(templates, t)
		}
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errTemplatesNotFound)
	}
	return templates, nil
}

func (tr *templateRepo) getAll(ctx context.Context) ([]models.Template, error) {
	op := "templateRepo.SearchPreparing.getAll"
	query := "SELECT id, title, description FROM templates"
	templates := []models.Template{}
	rows, err := tr.Storage.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		template := models.Template{}
		if err := rows.Scan(
			&template.Id,
			&template.Title,
			&template.Description,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		templates = append(templates, template)
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errTemplatesNotFound)
	}
	return templates, nil
}

func (tr *templateRepo) MustBulk(cfg config.SearchConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(int(time.Second)*cfg.Timeout))
	defer cancel()
	op := "templateRepo.SearchPreparing.Bulk"
	templates, err := tr.getAll(ctx)
	if err != nil {
		panic(fmt.Errorf("%s : %w", op, err))
	}
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: tr.SearchClient.Client,
		Index:  "templates",
	})
	if err != nil {
		panic(fmt.Errorf("%s : %w", op, err))
	}
	type doc struct {
		Id          string
		Title       string
		Description string
	}
	for _, t := range templates {
		d := doc{
			Id:          t.Id.String(),
			Title:       t.Title,
			Description: t.Description,
		}
		data, err := json.Marshal(d)
		if err != nil {
			panic(fmt.Errorf("%s : %w", op, err))
		}
		if err := bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: t.Id.String(),
			Body:       bytes.NewReader(data),
		}); err != nil {
			panic(fmt.Errorf("%s : %w", op, err))
		}
	}
	if err := bi.Close(ctx); err != nil {
		panic(fmt.Errorf("%s : %w\n stats: flushed - %d, failed - %d", op, err, bi.Stats().NumFlushed, bi.Stats().NumFailed))
	}
}
