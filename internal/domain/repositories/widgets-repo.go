package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/pkg/cache"
	"readmeow/pkg/search"
	"readmeow/pkg/storage"
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
	MustBulk()
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
	if err == cache.Empty {
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
	op := "widgerRepo.Search"
	mainQuery := types.Query{
		MultiMatch: &types.MultiMatchQuery{
			Query:     query,
			Fields:    []string{"title^5", "description^4", "description^3", "num_of_users^2", "likes"},
			Fuzziness: "AUTO",
		},
	}
	res, err := wr.SearchClient.Client.Search().Index("widgets").From(int(amount*page - amount)).Size(int(amount)).Request(&s.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: []types.Query{mainQuery},
			},
		},
		Source_: types.NewSourceIndex(),
	}).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ids := []string{}
	for _, hit := range res.Hits.Hits {
		ids = append(ids, *hit.Id_)
	}
	widgets, err := wr.getByIds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	return widgets, nil
}

func (wr *widgetRepo) getByIds(ctx context.Context, ids []string) ([]models.Widget, error) {
	op := "widgerRepo.SearchPreparing.getByIds"
	query := "SELECT * FROM widgets ANY($1)"
	widgets := []models.Widget{}
	rows, err := wr.Storage.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

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

func (wr *widgetRepo) getAll(ctx context.Context) ([]models.Widget, error) {
	op := "widgerRepo.SearchPreparing.getAll"
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

func (wr *widgetRepo) MustBulk() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	op := "widgerRepo.SearchPreparing.Bulk"
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
		panic(fmt.Errorf("%s : %w", op, err))
	}
}
