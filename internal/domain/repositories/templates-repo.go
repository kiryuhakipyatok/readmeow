package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"readmeow/internal/config"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories/helpers"
	"readmeow/pkg/cache"
	"readmeow/pkg/errs"
	"readmeow/pkg/search"
	"readmeow/pkg/storage"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9/esutil"
	s "github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
	"github.com/google/uuid"
)

type TemplateRepo interface {
	Create(ctx context.Context, template *models.Template) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	GetImage(ctx context.Context, id string) (string, error)
	Like(ctx context.Context, id, uid string) error
	Dislike(ctx context.Context, id, uid string) error
	FetchFavorite(ctx context.Context, id string, amount, page uint) ([]models.Template, error)
	Search(ctx context.Context, amount, page uint, query string, filter map[string]bool, sort map[string]string) ([]models.Template, error)
	MustBulk(ctx context.Context, cfg config.SearchConfig) error
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

func (tr *templateRepo) Create(ctx context.Context, template *models.Template) error {
	op := "templateRepo.Create"
	query := "INSERT INTO templates (id, owner_id, title, image, description, text, links, widgets,num_of_users, render_order, create_time, last_update_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)"
	qd := helpers.NewQueryData(ctx, tr.Storage, op, query, template.Id, template.OwnerId, template.Title, template.Image, template.Description, template.Text, template.Links, template.Widgets, template.NumOfUsers, template.RenderOrder, template.CreateTime, template.LastUpdateTime)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (tr *templateRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "templateRepo.Update"
	validFields := map[string]bool{
		"title":            true,
		"image":            true,
		"text":             true,
		"links":            true,
		"description":      true,
		"widgets":          true,
		"render_order":     true,
		"num_of_users":     true,
		"likes":            true,
		"last_update_time": true,
	}
	validValuesForLikesAndNumOfUsers := map[string]bool{
		"+": true,
		"-": true,
	}
	str := []string{}
	args := []any{}
	i := 1
	for k, v := range updates {
		if !validFields[k] {
			return errs.ErrInvalidFields(op)
		}
		if k == "likes" || k == "num_of_users" {
			val := v.(string)
			if !validValuesForLikesAndNumOfUsers[val] {
				return errs.ErrInvalidFields(op)
			}
			str = append(str, fmt.Sprintf(" %s = GREATEST(%s %s 1, 0)", k, k, val))
		} else {
			str = append(str, fmt.Sprintf(" %s = $%d", k, i))
			args = append(args, v)
			i++
		}
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE templates SET%s WHERE id = $%d", strings.Join(str, ","), i)
	fmt.Println(query)
	fmt.Println(args...)
	qd := helpers.NewQueryData(ctx, tr.Storage, op, query, args...)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	if err := tr.Cache.Redis.Del(ctx, id).Err(); err != nil {
		return errs.NewAppError(op, err)
	}
	return nil
}

func (tr *templateRepo) GetImage(ctx context.Context, id string) (string, error) {
	op := "templateRepo.GetImage"
	query := "SELECT image FROM templates WHERE id = $1"
	var url string
	qd := helpers.NewQueryData(ctx, tr.Storage, op, query, id)
	if err := qd.QueryRowWithTx(&url); err != nil {
		return "", err
	}
	return url, nil
}

func (tr *templateRepo) Like(ctx context.Context, id, uid string) error {
	op := "templateRepo.Like"
	query := "INSERT INTO favorite_templates (template_id, user_id) VALUES ($1,$2)"
	qd := helpers.NewQueryData(ctx, tr.Storage, op, query, id, uid)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (tr *templateRepo) Dislike(ctx context.Context, id, uid string) error {
	op := "templateRepo.Dislike"
	query := "DELETE FROM favorite_templates WHERE (template_id, user_id) = ($1,$2)"
	qd := helpers.NewQueryData(ctx, tr.Storage, op, query, id, uid)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (tr *templateRepo) Delete(ctx context.Context, id string) error {
	op := "templateRepo.Delete"
	query := "DELETE FROM templates WHERE id = $1"
	res, err := tr.Storage.Pool.Exec(ctx, query, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	if err := tr.Cache.Redis.Del(ctx, id).Err(); err != nil {
		return errs.NewAppError(op, err)
	}
	return nil
}

func (tr *templateRepo) Get(ctx context.Context, id string) (*models.Template, error) {
	op := "templateRepo.Get"
	template := &models.Template{}
	cachedTemplate, err := tr.Cache.Redis.Get(ctx, id).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedTemplate), template); err != nil {
			if err := tr.Cache.Redis.Del(ctx, id).Err(); err != nil {
				return nil, err
			}
			return nil, errs.NewAppError(op, err)
		}
		return template, nil
	}
	if err == cache.EMPTY {
		query := "SELECT * FROM templates WHERE id = $1"
		qd := helpers.NewQueryData(ctx, tr.Storage, op, query, id)
		if err := qd.QueryRowWithTx(template); err != nil {
			return nil, err
		}
	}
	if (template.NumOfUsers >= 20) || template.OwnerId == uuid.Nil {
		ttl := time.Hour * 24
		if template.NumOfUsers >= 100 {
			ttl = time.Hour * 48
		}
		cache, err := json.Marshal(template)
		if err != nil {
			return nil, errs.NewAppError(op, err)
		}
		if err := tr.Cache.Redis.Set(ctx, template.Id.String(), cache, ttl).Err(); err != nil {
			return nil, errs.NewAppError(op, err)
		}
	}
	return template, nil
}

func (tr *templateRepo) FetchFavorite(ctx context.Context, id string, amount, page uint) ([]models.Template, error) {
	op := "templateRepo.FetchFavorite"
	query := "SELECT t.* FROM templates t JOIN favorite_templates ft ON t.id=ft.template_id WHERE ft.user_id=$1 ORDER BY t.num_of_users DESC OFFSET $2 LIMIT $3"
	templates := []models.Template{}
	rows, err := tr.Storage.Pool.Query(ctx, query, id, amount*page-amount, amount)
	if err != nil {
		return nil, errs.NewAppError(op, err)
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
			&template.Likes,
			&template.RenderOrder,
			&template.CreateTime,
			&template.LastUpdateTime,
			&template.NumOfUsers,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (tr *templateRepo) Search(ctx context.Context, amount, page uint, query string, filter map[string]bool, sort map[string]string) ([]models.Template, error) {
	op := "templateRepo.Search"
	var mainQuery types.Query
	if query != "" {
		mainQuery = types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:     query,
				Fields:    []string{"Title^2", "Description"},
				Fuzziness: "AUTO",
			},
		}
	} else {
		mainQuery = types.Query{
			MatchAll: &types.MatchAllQuery{},
		}
	}

	sorts := []types.SortCombinations{}
	if len(sort) > 0 {
		validSortFields := map[string]bool{
			"Likes":          true,
			"NumOfUsers":     true,
			"LastUpdateTime": true,
		}
		validSortValues := map[string]bool{
			"desc": true,
			"asc":  true,
		}
		for k, v := range sort {
			if !validSortFields[k] {
				return nil, errs.ErrInvalidFields(op)
			} else if !validSortValues[strings.ToLower(v)] {
				return nil, errs.ErrInvalidValues(op)
			}
			order := &sortorder.Desc
			if v == "asc" {
				order = &sortorder.Asc
			}
			sorts = append(sorts, &types.SortOptions{
				SortOptions: map[string]types.FieldSort{
					k: {
						Order: order,
					},
				},
			})
		}
	}

	filters := []types.Query{}
	if len(filter) > 0 {
		validFilterFields := map[string]bool{
			"isOfficial": true,
		}
		for k, v := range filter {
			if !validFilterFields[k] {
				return nil, errs.ErrInvalidFields(op)
			}
			if v {
				filters = append(filters, types.Query{
					Term: map[string]types.TermQuery{
						"OwnerId.keyword": {Value: uuid.Nil.String()},
					},
				})
			} else {
				filters = append(filters, types.Query{
					Bool: &types.BoolQuery{
						MustNot: []types.Query{
							{
								Term: map[string]types.TermQuery{
									"OwnerId.keyword": {Value: uuid.Nil.String()},
								},
							},
						},
					},
				})
			}
		}
	}

	ptr := func(i int) *int {
		return &i
	}

	res, err := tr.SearchClient.Client.Search().Index("templates").Request(&s.Request{
		From: ptr(int(amount*page - amount)),
		Size: ptr(int(amount)),
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: []types.Query{mainQuery},
			},
		},
		Sort: sorts,
		PostFilter: &types.Query{
			Bool: &types.BoolQuery{
				Filter: filters,
			},
		},
		Source_: &types.SourceFilter{Includes: []string{"id"}},
	}).Do(ctx)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	ids := []string{}
	for _, hit := range res.Hits.Hits {
		if hit.Id_ != nil {
			ids = append(ids, *hit.Id_)
		}
	}
	if len(ids) == 0 {
		return nil, errs.ErrNotFound(op)
	}
	templates, err := tr.getByIds(ctx, ids)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}

	return templates, nil
}

func (tr *templateRepo) getByIds(ctx context.Context, ids []string) ([]models.Template, error) {
	fmt.Println(ids)
	op := "templateRepo.SearchPreparing.GetByIds"
	query := "SELECT * FROM templates WHERE id = ANY($1)"
	templates := make([]models.Template, 0, len(ids))
	rows, err := tr.Storage.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, errs.NewAppError(op, err)
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
			&template.Widgets,
			&template.Likes,
			&template.RenderOrder,
			&template.CreateTime,
			&template.LastUpdateTime,
			&template.NumOfUsers,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		byId[template.Id.String()] = template
	}
	if len(byId) == 0 {
		return nil, errs.ErrNotFound(op)
	}
	for _, id := range ids {
		if t, ok := byId[id]; ok {
			templates = append(templates, t)
		}
	}
	return templates, nil
}

func (tr *templateRepo) getAll(ctx context.Context) ([]models.Template, error) {
	op := "templateRepo.SearchPreparing.getAll"
	query := "SELECT id, owner_id, title, description, likes, num_of_users, last_update_time FROM templates"
	templates := []models.Template{}
	rows, err := tr.Storage.Pool.Query(ctx, query)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()

	for rows.Next() {
		template := models.Template{}
		if err := rows.Scan(
			&template.Id,
			&template.OwnerId,
			&template.Title,
			&template.Description,
			&template.Likes,
			&template.NumOfUsers,
			&template.LastUpdateTime,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		templates = append(templates, template)
	}
	if len(templates) == 0 {
		return nil, errs.ErrNotFound(op)
	}
	return templates, nil
}

func (tr *templateRepo) MustBulk(ctx context.Context, cfg config.SearchConfig) error {
	op := "templateRepo.SearchPreparing.Bulk"
	templates, err := tr.getAll(ctx)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: tr.SearchClient.Client,
		Index:  "templates",
	})
	if err != nil {
		return errs.NewAppError(op, err)

	}
	type doc struct {
		Id             string
		OwnerId        string
		Title          string
		Description    string
		Likes          uint32
		NumOfUsers     uint32
		LastUpdateTime time.Time
	}
	for _, t := range templates {
		d := doc{
			Id:             t.Id.String(),
			OwnerId:        t.OwnerId.String(),
			Title:          t.Title,
			Description:    t.Description,
			NumOfUsers:     t.NumOfUsers,
			Likes:          t.Likes,
			LastUpdateTime: t.LastUpdateTime,
		}
		data, err := json.Marshal(d)
		if err != nil {
			return errs.NewAppError(op, err)
		}
		if err := bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: t.Id.String(),
			Body:       bytes.NewReader(data),
		}); err != nil {
			return errs.NewAppError(op, err)
		}
	}
	if err := bi.Close(ctx); err != nil {
		return errs.NewAppError(op, fmt.Errorf("%w, stats: flushed - %d, failed - %d", err, bi.Stats().NumFlushed, bi.Stats().NumFailed))
	}
	return nil
}
