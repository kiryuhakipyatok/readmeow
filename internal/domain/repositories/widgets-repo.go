package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
)

type WidgetRepo interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error)
	Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error)
	Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error)
	Like(ctx context.Context, uid, id string) error
	Dislike(ctx context.Context, uid, id string) error
	FetchFavorite(ctx context.Context, id string) ([]models.Widget, error)
	GetByIds(ctx context.Context, ids []string) ([]models.Widget, error)
	Update(ctx context.Context, updates map[string]any, id string) error
	MustBulk(ctx context.Context, cfg config.SearchConfig) error
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

func (wr *widgetRepo) Get(ctx context.Context, id string) (*models.Widget, error) {
	op := "widgetRepo.Get"
	widget := &models.Widget{}
	cachedWidget, err := wr.Cache.Redis.Get(ctx, id).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedWidget), widget); err != nil {
			if err := wr.Cache.Redis.Del(ctx, id).Err(); err != nil {
				return nil, err
			}
			return nil, errs.NewAppError(op, err)
		}
		return widget, nil
	}
	if err == cache.EMPTY {
		query := "SELECT w.*, COUNT(fw.widget_id) as likes FROM widgets w LEFT JOIN favorite_widgets fw ON w.id=fw.widget_id WHERE w.id = $1 GROUP BY w.id"
		qd := helpers.NewQueryData(ctx, wr.Storage, op, query, id)
		if err := qd.QueryRowWithTx(widget); err != nil {
			return nil, err
		}
	}
	cache, err := json.Marshal(widget)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	ttl := time.Hour * 24
	if widget.NumOfUsers >= 100 {
		ttl = time.Hour * 48
	}
	if err := wr.Cache.Redis.Set(ctx, widget.Id.String(), cache, ttl).Err(); err != nil {
		return nil, errs.NewAppError(op, err)
	}
	return widget, nil
}

func (wr *widgetRepo) Fetch(ctx context.Context, amount, page uint) ([]models.Widget, error) {
	op := "widgetRepo.Fetch"
	query := "SELECT w.*, COUNT(fw.widget_id) as likes FROM widgets w LEFT JOIN favorite_widgets fw ON w.id=fw.widget_id GROUP BY w.id ORDER BY likes DESC OFFSET $1 LIMIT $2"
	rows, err := wr.Storage.Pool.Query(ctx, query, amount*page-amount, amount)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
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
			&widget.NumOfUsers,
			&widget.Tags,
			&widget.Likes,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		widgets = append(widgets, widget)
	}
	return widgets, nil
}

func (wr *widgetRepo) FetchFavorite(ctx context.Context, id string) ([]models.Widget, error) {
	op := "widgetRepo.FetchFavorite"
	query := "SELECT w.*, COUNT(fw.widget_id) as likes FROM widgets w JOIN favorite_widgets fw ON w.id=fw.widget_id WHERE fw.user_id=$1 GROUP BY w.id"
	widgets := []models.Widget{}
	rows, err := wr.Storage.Pool.Query(ctx, query, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
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
			&widget.NumOfUsers,
			&widget.Tags,
			&widget.Likes,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		widgets = append(widgets, widget)
	}
	return widgets, nil
}

func (wr *widgetRepo) Like(ctx context.Context, uid, id string) error {
	op := "widgetRepo.Like"
	query := "INSERT INTO favorite_widgets (widget_id, user_id) VALUES($1,$2)"
	qd := helpers.NewQueryData(ctx, wr.Storage, op, query, id, uid)
	if err := qd.InsertWithTx(); err != nil {
		return err
	}
	return nil
}

func (wr *widgetRepo) Dislike(ctx context.Context, uid, id string) error {
	op := "widgetRepo.Dislike"
	query := "DELETE FROM favorite_widgets WHERE (widget_id,user_id)=($1,$2)"
	qd := helpers.NewQueryData(ctx, wr.Storage, op, query, id, uid)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	return nil
}

func (wr *widgetRepo) Sort(ctx context.Context, amount, page uint, field, dest string) ([]models.Widget, error) {
	op := "widgetRepo.Sort"
	validFields := map[string]bool{
		"likes":        true,
		"num_of_users": true,
	}
	if !validFields[field] {
		return nil, errs.ErrInvalidFields(op)
	}
	if dest != "DESC" && dest != "ASC" {
		dest = "DESC"
	}
	query := fmt.Sprintf("SELECT w.*, COUNT(fw.widget_id) as likes FROM widgets w LEFT JOIN favorite_widgets fw ON w.id=fw.widget_id GROUP BY w.id ORDER BY %s %s OFFSET $1 LIMIT $2", field, dest)
	rows, err := wr.Storage.Pool.Query(ctx, query, amount*page-amount, amount)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
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
			&widget.NumOfUsers,
			&widget.Tags,
			&widget.Likes,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		widgets = append(widgets, widget)
	}
	return widgets, nil
}

func (wr *widgetRepo) Search(ctx context.Context, amount, page uint, query string) ([]models.Widget, error) {
	op := "widgetRepo.Search"
	mainQuery := types.Query{
		MultiMatch: &types.MultiMatchQuery{
			Query:     query,
			Fields:    []string{"Title^3", "Type^2", "Description"},
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
		return nil, errs.NewAppError(op, err)
	}
	ids := []string{}
	for _, hit := range res.Hits.Hits {
		if hit.Id_ != nil {
			ids = append(ids, *hit.Id_)
		}
	}
	widgets, err := wr.GetByIds(ctx, ids)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	return widgets, nil
}

func (wr *widgetRepo) GetByIds(ctx context.Context, ids []string) ([]models.Widget, error) {
	op := "widgetRepo.SearchPreparing.GetByIds"
	query := "SELECT w.*, COUNT(fw.widget_id) as likes FROM widgets w LEFT JOIN favorite_widgets fw ON w.id=fw.widget_id WHERE w.id = ANY($1) GROUP BY w.id"
	widgets := make([]models.Widget, 0, len(ids))
	byId := map[string]models.Widget{}
	if tx, ok := storage.GetTx(ctx); ok {
		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound()) {
				return nil, errs.ErrNotFound(op)
			}
			return nil, errs.NewAppError(op, err)
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
				&widget.NumOfUsers,
				&widget.Tags,
				&widget.Likes,
			); err != nil {
				return nil, errs.NewAppError(op, err)
			}
			byId[widget.Id.String()] = widget
		}
	} else {
		rows, err := wr.Storage.Pool.Query(ctx, query, ids)
		if err != nil {
			return nil, errs.NewAppError(op, err)
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
				&widget.NumOfUsers,
				&widget.Tags,
				&widget.Likes,
			); err != nil {
				return nil, errs.NewAppError(op, err)
			}
			byId[widget.Id.String()] = widget
		}
	}
	for _, id := range ids {
		if w, ok := byId[id]; ok {
			widgets = append(widgets, w)
		}
	}
	return widgets, nil
}

func (wr *widgetRepo) getAll(ctx context.Context) ([]models.Widget, error) {
	op := "widgetRepo.SearchPreparing.getAll"
	query := "SELECT id, title, description, type FROM widgets"
	widgets := []models.Widget{}
	rows, err := wr.Storage.Pool.Query(ctx, query)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()

	for rows.Next() {
		widget := models.Widget{}
		if err := rows.Scan(
			&widget.Id,
			&widget.Title,
			&widget.Description,
			&widget.Type,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		widgets = append(widgets, widget)
	}
	return widgets, nil
}

func (wr *widgetRepo) MustBulk(ctx context.Context, cfg config.SearchConfig) error {
	op := "widgetRepo.SearchPreparing.Bulk"
	widgets, err := wr.getAll(ctx)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: wr.SearchClient.Client,
		Index:  "widgets",
	})
	if err != nil {
		return errs.NewAppError(op, err)
	}
	type doc struct {
		Id          string
		Title       string
		Description string
		Type        string
	}
	for _, w := range widgets {
		d := doc{
			Id:          w.Id.String(),
			Title:       w.Title,
			Description: w.Description,
			Type:        w.Type,
		}
		data, err := json.Marshal(d)
		if err != nil {
			return errs.NewAppError(op, err)
		}
		if err := bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: w.Id.String(),
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

func (wr *widgetRepo) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "widgetRepo.Update"
	validFields := map[string]bool{
		"num_of_users": true,
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
	query := fmt.Sprintf("UPDATE widgets SET%s WHERE id = $%d", strings.Join(str, ","), i)
	qd := helpers.NewQueryData(ctx, wr.Storage, op, query, args...)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	if err := wr.Cache.Redis.Del(ctx, id).Err(); err != nil {
		return errs.NewAppError(op, err)
	}
	return nil
}
