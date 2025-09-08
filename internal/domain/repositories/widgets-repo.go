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
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
)

type WidgetRepo interface {
	Get(ctx context.Context, id string) (*models.Widget, error)
	Search(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error)
	Like(ctx context.Context, uid, id string) error
	Dislike(ctx context.Context, uid, id string) error
	FetchFavorite(ctx context.Context, id string, amount, page uint) ([]models.Widget, error)
	GetByIds(ctx context.Context, ids []string) ([]models.Widget, error)
	Update(ctx context.Context, updates map[string]string, id string) error
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
		query := "SELECT * FROM widgets WHERE id = $1"
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

func (wr *widgetRepo) FetchFavorite(ctx context.Context, id string, amount, page uint) ([]models.Widget, error) {
	op := "widgetRepo.FetchFavorite"
	query := "SELECT w.* FROM widgets w JOIN favorite_widgets fw ON w.id=fw.widget_id WHERE fw.user_id=$1 ORDER BY w.num_of_users DESC OFFSET $2 LIMIT $3"
	widgets := []models.Widget{}
	rows, err := wr.Storage.Pool.Query(ctx, query, id, amount*page-amount, amount)
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
			&widget.Tags,
			&widget.Link,
			&widget.Likes,
			&widget.NumOfUsers,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		widgets = append(widgets, widget)
	}
	if len(widgets) == 0 {
		return []models.Widget{}, nil
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

func (wr *widgetRepo) Search(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
	op := "widgetRepo.Search"
	var mainQuery types.Query
	if query != "" {
		mainQuery = types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:     query,
				Fields:    []string{"Title^3", "Type^2", "Description"},
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
			"Likes":      true,
			"NumOfUsers": true,
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
			"Tags":  true,
			"Types": true,
		}
		for k := range filter {
			if !validFilterFields[k] {
				return nil, errs.ErrInvalidFields(op)
			}
		}

		if tags, ok := filter["Tags"]; ok && len(tags) > 0 {
			tagsQuery := []types.Query{}
			for _, t := range tags {
				tagsQuery = append(tagsQuery, types.Query{
					Exists: &types.ExistsQuery{
						Field: fmt.Sprintf("Tags.%s", t),
					},
				})
			}
			filters = append(filters, types.Query{
				Bool: &types.BoolQuery{
					Should:             tagsQuery,
					MinimumShouldMatch: 1,
				},
			})
		}

		if typs, ok := filter["Types"]; ok && len(typs) > 0 {
			typesQuery := []types.Query{}
			for _, t := range typs {
				typesQuery = append(typesQuery, types.Query{
					Term: map[string]types.TermQuery{
						"Type.keyword": {Value: t},
					},
				})
			}
			filters = append(filters, types.Query{
				Bool: &types.BoolQuery{
					Should:             typesQuery,
					MinimumShouldMatch: 1,
				},
			})
		}

	}

	ptr := func(i int) *int {
		return &i
	}

	res, err := wr.SearchClient.Client.Search().Index("widgets").Request(&s.Request{
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
		return []models.Widget{}, nil
	}
	widgets, err := wr.GetByIds(ctx, ids)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	return widgets, nil
}

func (wr *widgetRepo) GetByIds(ctx context.Context, ids []string) ([]models.Widget, error) {
	op := "widgetRepo.SearchPreparing.GetByIds"
	query := "SELECT * FROM widgets WHERE id = ANY($1)"
	widgets := make([]models.Widget, 0, len(ids))
	byId := map[string]models.Widget{}
	if tx, ok := storage.GetTx(ctx); ok {
		rows, err := tx.Query(ctx, query, ids)
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
				&widget.Tags,
				&widget.Link,
				&widget.Likes,
				&widget.NumOfUsers,
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
				&widget.Tags,
				&widget.Link,
				&widget.Likes,
				&widget.NumOfUsers,
			); err != nil {
				return nil, errs.NewAppError(op, err)
			}
			byId[widget.Id.String()] = widget
		}
	}
	if len(byId) == 0 {
		return []models.Widget{}, nil
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
	query := "SELECT id, title, description, type, likes, num_of_users,tags FROM widgets"
	widgets := []models.Widget{}
	rows, err := wr.Storage.Pool.Query(ctx, query)
	if err != nil {
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
			&widget.Likes,
			&widget.NumOfUsers,
			&widget.Tags,
		); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		widgets = append(widgets, widget)
	}
	if len(widgets) == 0 {
		return []models.Widget{}, nil
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
		Likes       uint32
		NumOfUsers  uint32
		Tags        map[string]any
	}
	for _, w := range widgets {
		d := doc{
			Id:          w.Id.String(),
			Title:       w.Title,
			Description: w.Description,
			Type:        w.Type,
			Likes:       w.Likes,
			Tags:        w.Tags,
			NumOfUsers:  w.NumOfUsers,
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

func (wr *widgetRepo) Update(ctx context.Context, updates map[string]string, id string) error {
	op := "widgetRepo.Update"
	validFields := map[string]bool{
		"num_of_users": true,
		"likes":        true,
	}
	validValues := map[string]bool{
		"+": true,
		"-": true,
	}
	str := []string{}
	for k, v := range updates {
		if !validFields[k] || !validValues[v] {
			return errs.ErrInvalidFields(op)
		}
		str = append(str, fmt.Sprintf(" %s = GREATEST(%s %s 1, 0)", k, k, v))
	}
	query := fmt.Sprintf("UPDATE widgets SET%s WHERE id = $1", strings.Join(str, ","))
	qd := helpers.NewQueryData(ctx, wr.Storage, op, query, id)
	if err := qd.DeleteOrUpdateWithTx(); err != nil {
		return err
	}
	if err := wr.Cache.Redis.Del(ctx, id).Err(); err != nil {
		return errs.NewAppError(op, err)
	}
	return nil
}
