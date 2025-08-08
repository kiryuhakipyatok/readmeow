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

	"github.com/elastic/go-elasticsearch/v8/esutil"
	s "github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

type WidgetRepo interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error)
	Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error)
	GetByIds(ctx context.Context, ids []string) ([]models.Widget, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	MustBulk(cfg *config.SearchConfig)
}

type widgetRepo struct {
	Storage      *storage.Storage
	Cache        *cache.Cache
	SearchClient *search.SearchClient
}

func NewWidgetRepo(s *storage.Storage, c *cache.Cache, sc *search.SearchClient) WidgetRepo {
	return &widgetRepo{
		Storage:      s,
		Cache:        c,
		SearchClient: sc,
	}
}

var (
	errWidgetNotFound  = errors.New("widget not found")
	errWidgetsNotFound = errors.New("widgets not found")
)

func (wr *widgetRepo) Get(ctx context.Context, id string) (*models.Widget, error) {
	op := "widgetRepo.Get"
	widget := &models.Widget{}
	cachedWidget, err := wr.Cache.Redis.Get(ctx, id).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedWidget), widget); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		return widget, nil
	}
	if err == cache.EMPTY {
		query := "SELECT * FROM widgets WHERE id = $1"
		if tx, ok := storage.GetTx(ctx); ok {
			if err := tx.QueryRow(ctx, query, id).Scan(
				&widget.Id,
				&widget.Title,
				&widget.Image,
				&widget.Description,
				&widget.Type,
				&widget.Link,
				&widget.Likes,
				&widget.NumOfUsers,
			); err != nil {
				if errors.Is(err, storage.ErrNotFound()) {
					return nil, fmt.Errorf("%s : %w", op, errWidgetNotFound)
				}
				return nil, fmt.Errorf("%s : %w", op, err)
			}
		} else {
			if err := wr.Storage.Pool.QueryRow(ctx, query, id).Scan(
				&widget.Id,
				&widget.Title,
				&widget.Image,
				&widget.Description,
				&widget.Type,
				&widget.Link,
				&widget.Likes,
				&widget.NumOfUsers,
			); err != nil {
				if errors.Is(err, storage.ErrNotFound()) {
					return nil, fmt.Errorf("%s : %w", op, errWidgetNotFound)
				}
				return nil, fmt.Errorf("%s : %w", op, err)
			}
		}
	}
	cache, err := json.Marshal(widget)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	if err := wr.Cache.Redis.Set(ctx, widget.Id.String(), cache, time.Hour*24).Err(); err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return widget, nil
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
		if err := rows.Scan(
			&widget.Id,
			&widget.Title,
			&widget.Image,
			&widget.Description,
			&widget.Type,
			&widget.Link,
			&widget.Likes,
			&widget.NumOfUsers,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		widgets = append(widgets, widget)
	}
	if len(widgets) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errWidgetsNotFound)
	}
	return widgets, nil
}

func (wr *widgetRepo) Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error) {
	op := "widgetRepo.Filter"
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
		if err := rows.Scan(
			&widget.Id,
			&widget.Title,
			&widget.Image,
			&widget.Description,
			&widget.Type,
			&widget.Link,
			&widget.Likes,
			&widget.NumOfUsers,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		widgets = append(widgets, widget)
	}
	if len(widgets) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errWidgetsNotFound)
	}
	return widgets, nil
}

func (wr *widgetRepo) Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error) {
	op := "widgetRepo.Search"
	mainQuery := types.Query{
		MultiMatch: &types.MultiMatchQuery{
			Query:     query,
			Fields:    []string{"title^5", "type^4", "description^3", "num_of_users^2", "likes"},
			Fuzziness: "AUTO",
		},
	}
	res, err := wr.SearchClient.Client.Search().Index("widgets").From(int(amount*page - amount)).Size(int(amount)).Request(&s.Request{
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
		ids = append(ids, *hit.Id_)
	}
	widgets, err := wr.GetByIds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return widgets, nil
}

func (wr *widgetRepo) GetByIds(ctx context.Context, ids []string) ([]models.Widget, error) {
	op := "widgetRepo.SearchPreparing.getByIds"
	query := "SELECT * FROM widgets WHERE id = ANY($1)"
	widgets := make([]models.Widget, 0, len(ids))
	rows, err := wr.Storage.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()
	byId := map[string]models.Widget{}
	for rows.Next() {
		widget := models.Widget{}
		if err := rows.Scan(
			&widget.Id,
			&widget.Title,
			&widget.Image,
			&widget.Description,
			&widget.Type,
			&widget.Link,
			&widget.Likes,
			&widget.NumOfUsers,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		byId[widget.Id.String()] = widget
	}
	for _, id := range ids {
		if w, ok := byId[id]; ok {
			widgets = append(widgets, w)
		}
	}
	if len(widgets) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errWidgetsNotFound)
	}
	return widgets, nil
}

func (wr *widgetRepo) getAll(ctx context.Context) ([]models.Widget, error) {
	op := "widgetRepo.SearchPreparing.getAll"
	query := "SELECT id, title, description, type, likes, num_of_users FROM widgets"
	widgets := []models.Widget{}
	rows, err := wr.Storage.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		widget := models.Widget{}
		if err := rows.Scan(
			&widget.Id,
			&widget.Title,
			&widget.Description,
			&widget.Type,
			&widget.Likes,
			&widget.NumOfUsers,
		); err != nil {
			return nil, fmt.Errorf("%s : %w", op, err)
		}
		widgets = append(widgets, widget)
	}
	if len(widgets) == 0 {
		return nil, fmt.Errorf("%s : %w", op, errWidgetsNotFound)
	}
	return widgets, nil
}

func (wr *widgetRepo) MustBulk(cfg *config.SearchConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout))
	defer cancel()
	op := "widgetRepo.SearchPreparing.Bulk"
	widgets, err := wr.getAll(ctx)
	if err != nil {
		panic(fmt.Errorf("%s : %w", op, err))
	}
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: wr.SearchClient.Client,
		Index:  "widgets",
	})
	if err != nil {
		panic(fmt.Errorf("%s : %w", op, err))
	}
	type doc struct {
		Id          string
		Title       string
		Description string
		Type        string
		Likes       uint16
		NumOfUsers  uint16
	}
	for _, w := range widgets {
		d := doc{
			Id:          w.Id.String(),
			Title:       w.Title,
			Description: w.Description,
			Type:        w.Type,
			Likes:       w.Likes,
			NumOfUsers:  w.NumOfUsers,
		}
		data, err := json.Marshal(d)
		if err != nil {
			panic(fmt.Errorf("%s : %w", op, err))
		}
		if err := bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: w.Id.String(),
			Body:       bytes.NewReader(data),
		}); err != nil {
			panic(fmt.Errorf("%s : %w", op, err))
		}
	}
	if err := bi.Close(ctx); err != nil {
		panic(fmt.Errorf("%s : %w\n stats: flushed - %d, failed - %d", op, err, bi.Stats().NumFlushed, bi.Stats().NumFailed))
	}
}

func (wr *widgetRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "widgetRepo.Update"
	validFields := map[string]bool{
		"likes":        true,
		"num_of_users": true,
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
	query := fmt.Sprintf("UPDATE widgets SET %s WHERE id = $%d", strings.Join(str, ","), i)
	if tx, ok := storage.GetTx(ctx); ok {
		res, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("%s : %w", op, errWidgetNotFound)
		}
	} else {
		res, err := wr.Storage.Pool.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("%s : %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("%s : %w", op, errWidgetNotFound)
		}
	}
	widget, err := wr.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	cache, err := json.Marshal(widget)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	ttl, err := wr.Cache.Redis.TTL(ctx, widget.Id.String()).Result()
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	if err := wr.Cache.Redis.Set(ctx, widget.Id.String(), cache, ttl).Err(); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	return nil
}
