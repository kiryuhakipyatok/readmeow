package services

import (
	"context"
	"fmt"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type TemplateServ interface {
	Create(ctx context.Context, oid, title, image, description, order string, text, links, widgets []string) error
	Update(ctx context.Context, fields map[string]any, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.Template, error)
	Fetch(ctx context.Context, amount, page uint) ([]models.Template, error)
	Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error)
	Search(ctx context.Context, amount, page uint, query string) ([]models.Template, error)
	Like(ctx context.Context, id string) error
	Dislike(ctx context.Context, id string) error
}

type templateServ struct {
	TemplateRepo repositories.TemplateRepo
	UserRepo     repositories.UserRepo
	WidgetRepo   repositories.WidgetRepo
	Transactor   storage.Transactor
	Logger       *logger.Logger
}

func NewTemplateServ(tr repositories.TemplateRepo, ur repositories.UserRepo, wr repositories.WidgetRepo, t storage.Transactor, l *logger.Logger) TemplateServ {
	return &templateServ{
		TemplateRepo: tr,
		UserRepo:     ur,
		Transactor:   t,
		WidgetRepo:   wr,
		Logger:       l,
	}
}

func (ts *templateServ) Create(ctx context.Context, oid, title, image, description, order string, text, links, widgets []string) error {
	op := "templateServ.Create"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("creating template")
	_, err := ts.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := ts.UserRepo.Get(ctx, oid)
		if err != nil {
			ts.Logger.Log.Error("failed to get user", logger.Err(err))
			return nil, err
		}
		template := &models.Template{
			Id:             uuid.New(),
			OwnerId:        user.Id,
			Title:          title,
			Image:          image,
			Description:    description,
			Text:           text,
			Links:          links,
			Widgets:        widgets,
			Likes:          0,
			NumOfUsers:     0,
			Order:          order,
			CreateTime:     time.Now().Unix(),
			LastUpdateTime: time.Now().Unix(),
		}
		if len(widgets) != 0 {
			widgetsData, err := ts.WidgetRepo.GetByIds(ctx, widgets)
			if err != nil {
				ts.Logger.Log.Error("failed to fetch widgets", logger.Err(err))
				return nil, err
			}

			for _, w := range widgetsData {
				updatedNumOfUsers := w.NumOfUsers + 1
				update := map[string]any{
					"num_of_users": updatedNumOfUsers,
				}
				if err := ts.WidgetRepo.Update(ctx, update, w.Id.String()); err != nil {
					ts.Logger.Log.Error("failed to update widget info", logger.Err(err))
					return nil, err
				}
			}

		}
		if err := ts.TemplateRepo.Create(ctx, template); err != nil {
			ts.Logger.Log.Error("failed to create template", logger.Err(err))
			return nil, err
		}
		numOfTemplates := user.NumOfTemplates + 1
		update := map[string]any{
			"num_of_templates": numOfTemplates,
		}
		if err := ts.UserRepo.Update(ctx, update, user.Id.String()); err != nil {
			ts.Logger.Log.Error("failed to update user info", logger.Err(err))
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		ts.Logger.Log.Error("failed to create template", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("template created successfully")
	return nil
}

func (ts *templateServ) Update(ctx context.Context, fields map[string]any, id string) error {
	op := "templateServ.Update"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("updating template")
	if err := ts.TemplateRepo.Update(ctx, fields, id); err != nil {
		ts.Logger.Log.Error("failed to update template", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("template updated successfully")
	return nil
}

func (ts *templateServ) Delete(ctx context.Context, id string) error {
	op := "templateServ.Delete"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("deleting template")
	if err := ts.TemplateRepo.Delete(ctx, id); err != nil {
		ts.Logger.Log.Error("failed to delete template", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("template deleted successfully")
	return nil
}

func (ts *templateServ) Get(ctx context.Context, id string) (*models.Template, error) {
	op := "templateServ.Get"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("receiving template")
	template, err := ts.TemplateRepo.Get(ctx, id)
	if err != nil {
		ts.Logger.Log.Error("failed to receive template", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("template received successfully")
	return template, nil
}

func (ts *templateServ) Fetch(ctx context.Context, amount, page uint) ([]models.Template, error) {
	op := "templateServ.Fetch"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("fetching templates")
	templates, err := ts.TemplateRepo.Fetch(ctx, amount, page)
	if err != nil {
		ts.Logger.Log.Error("failed to fetch templates", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("template fetched successfully")
	return templates, nil
}

func (ts *templateServ) Sort(ctx context.Context, amount, page uint, dest, field string) ([]models.Template, error) {
	op := "templateServ.Sort"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("fetching sorted templates")
	templates, err := ts.TemplateRepo.Sort(ctx, amount, page, dest, field)
	if err != nil {
		ts.Logger.Log.Error("failed to fetch sorted templates", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("sorted templates fetched successfully")
	return templates, nil
}

func (ts *templateServ) Search(ctx context.Context, amount, page uint, query string) ([]models.Template, error) {
	op := "templateServ.Search"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("fetchin searched templates")
	templates, err := ts.TemplateRepo.Search(ctx, amount, page, query)
	if err != nil {
		ts.Logger.Log.Error("failed to fetch searched templates", logger.Err(err))
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	ts.Logger.Log.Info("searched templates fetched successfully")
	return templates, nil
}

func (ts *templateServ) Like(ctx context.Context, id string) error {
	op := "templateServ.Like"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("liking template")
	template, err := ts.TemplateRepo.Get(ctx, id)
	if err != nil {
		ts.Logger.Log.Error("failed to get widget", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	updatedLikes := template.Likes + 1
	update := map[string]any{
		"likes": updatedLikes,
	}
	if err := ts.TemplateRepo.Update(ctx, update, template.Id.String()); err != nil {
		ts.Logger.Log.Error("failed to update template info", logger.Err(err))
		return err
	}
	ts.Logger.Log.Info("template liked successfully")
	return nil
}

func (ts *templateServ) Dislike(ctx context.Context, id string) error {
	op := "templateServ.Dislike"
	ts.Logger.AddOp(op)
	ts.Logger.Log.Info("disliking template")
	widget, err := ts.TemplateRepo.Get(ctx, id)
	if err != nil {
		ts.Logger.Log.Error("failed to get template", logger.Err(err))
		return fmt.Errorf("%s : %w", op, err)
	}
	updatedLikes := widget.Likes - 1
	update := map[string]any{
		"likes": updatedLikes,
	}
	if err := ts.TemplateRepo.Update(ctx, update, widget.Id.String()); err != nil {
		ts.Logger.Log.Error("failed to update template info", logger.Err(err))
		return err
	}
	ts.Logger.Log.Info("template disliked successfully")
	return nil
}
