package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"readmeow/internal/domain/models"
	"readmeow/internal/domain/repositories"
	"readmeow/internal/dto"
	"readmeow/pkg/cloudstorage"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	"readmeow/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type ReadmeServ interface {
	Create(ctx context.Context, tid, oid, title string, image *multipart.FileHeader, text, links, order []string, widgets []map[string]string) error
	Delete(ctx context.Context, id, uid string) error
	Update(ctx context.Context, updates map[string]any, id string) error
	Get(ctx context.Context, id string) (*models.Readme, error)
	FetchByUser(ctx context.Context, amount, page uint, uid string) ([]dto.ReadmeResponse, error)
}

type readmeServ struct {
	ReadmeRepo   repositories.ReadmeRepo
	UserRepo     repositories.UserRepo
	TemplateRepo repositories.TemplateRepo
	WidgetRepo   repositories.WidgetRepo
	Transactor   storage.Transactor
	CloudStorage cloudstorage.CloudStorage
	Logger       *logger.Logger
}

func NewReadmeServ(rr repositories.ReadmeRepo, ur repositories.UserRepo, tr repositories.TemplateRepo, wr repositories.WidgetRepo, t storage.Transactor, cs cloudstorage.CloudStorage, l *logger.Logger) ReadmeServ {
	return &readmeServ{
		ReadmeRepo:   rr,
		UserRepo:     ur,
		TemplateRepo: tr,
		WidgetRepo:   wr,
		Logger:       l,
		Transactor:   t,
		CloudStorage: cs,
	}
}

func (rs *readmeServ) Create(ctx context.Context, tid, oid, title string, image *multipart.FileHeader, text, links, order []string, widgets []map[string]string) error {
	op := "readmeServ.Create"
	log := rs.Logger.AddOp(op)
	log.Log.Info("creating readme")
	_, err := rs.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := rs.UserRepo.Get(c, oid)
		if err != nil {
			return nil, err
		}
		if tid == "" {
			tid = baseTemplateId.String()
		}
		template, err := rs.TemplateRepo.Get(c, tid)
		if err != nil {
			return nil, err
		}
		updateT := map[string]any{
			"num_of_users": "+",
		}
		if err := rs.TemplateRepo.Update(c, updateT, template.Id.String()); err != nil {
			return nil, err
		}
		id := uuid.New()

		if len(widgets) != 0 {
			keys := make([]string, 0, len(widgets))
			for _, w := range widgets {
				for k := range w {
					keys = append(keys, k)
				}
			}
			widgetsData, err := rs.WidgetRepo.GetByIds(c, keys)
			if err != nil {
				return nil, err
			}

			for _, w := range widgetsData {
				updateW := map[string]string{
					"num_of_users": "+",
				}
				if err := rs.WidgetRepo.Update(c, updateW, w.Id.String()); err != nil {
					return nil, err
				}
			}

		}
		updateR := map[string]any{
			"num_of_readmes": "+",
		}
		if err := rs.UserRepo.Update(c, updateR, user.Id.String()); err != nil {
			return nil, err
		}

		file, err := image.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()
		folder := "readmes"
		now := time.Now()
		unow := now.Unix()
		filename := fmt.Sprintf("%s-%d", id, unow)
		url, pid, err := rs.CloudStorage.UploadImage(c, file, filename, folder)
		if err != nil {
			return nil, err
		}

		readme := &models.Readme{
			Id:             id,
			OwnerId:        user.Id,
			TemplateId:     template.Id,
			Title:          title,
			Image:          url,
			Text:           text,
			Links:          links,
			Widgets:        widgets,
			RenderOrder:    order,
			CreateTime:     now,
			LastUpdateTime: now,
		}

		if err := rs.ReadmeRepo.Create(c, readme); err != nil {
			if cerr := rs.CloudStorage.DeleteImage(c, pid); cerr != nil {
				return nil, fmt.Errorf("%w : %w", err, cerr)
			}
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		log.Log.Error("failed to create new readme", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Log.Info("new readme created successfully")
	return nil
}

func (rs *readmeServ) Delete(ctx context.Context, id, uid string) error {
	op := "readmeServ"
	log := rs.Logger.AddOp(op)
	log.Log.Info("deleting readme")
	if _, err := rs.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		user, err := rs.UserRepo.Get(c, uid)
		if err != nil {
			return nil, err
		}
		readme, err := rs.ReadmeRepo.Get(c, id)
		if err != nil {
			return nil, err
		}
		if readme.OwnerId != user.Id {
			err := errors.New("readme owner id and user id are not equal")
			return nil, err
		}
		widgets := make(map[string]struct{})
		for _, w := range readme.Widgets {
			for id := range w {
				widgets[id] = struct{}{}
			}
		}
		wupd := map[string]string{
			"num_of_users": "-",
		}
		for wid := range widgets {
			if err := rs.WidgetRepo.Update(c, wupd, wid); err != nil {
				return nil, err
			}
		}
		tupd := map[string]any{
			"num_of_users": "-",
		}
		if err := rs.TemplateRepo.Update(c, tupd, readme.TemplateId.String()); err != nil {
			return nil, err
		}

		if err := rs.ReadmeRepo.Delete(c, id); err != nil {
			return nil, err
		}
		pid, err := rs.CloudStorage.GetPIdFromURL(readme.Image)
		if err != nil {
			return nil, err
		}
		if err := rs.CloudStorage.DeleteImage(c, pid); err != nil {
			return nil, err
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("failed to delete readme", logger.Err(err))
		return errs.NewAppError(op, err)
	}

	log.Log.Info("readme deleted successfully")
	return nil
}

func (rs *readmeServ) Update(ctx context.Context, updates map[string]any, id string) error {
	op := "readmeServ.Update"
	log := rs.Logger.AddOp(op)
	log.Log.Info("updating readme")
	if _, err := rs.Transactor.WithinTransaction(ctx, func(c context.Context) (any, error) {
		fileAnyH, fOk := updates["image"]
		widgs, wOk := updates["widgets"]
		var (
			newPid string
			oldURL string
		)
		now := time.Now()
		if fOk || wOk {
			readme, err := rs.Get(c, id)
			if err != nil {
				return nil, err
			}

			if fOk {
				oldURL = readme.Image
				fileH := fileAnyH.(*multipart.FileHeader)
				file, err := fileH.Open()
				if err != nil {
					return nil, err
				}
				defer file.Close()
				folder := "readmes"
				unow := now.Unix()
				filename := fmt.Sprintf("%s-%d", id, unow)
				var url string
				url, newPid, err = rs.CloudStorage.UploadImage(c, file, filename, folder)
				if err != nil {
					return nil, err
				}
				updates["image"] = url
			}
			if wOk {
				newWidgets := widgs.([]map[string]string)
				nwids := make(map[string]struct{}, len(newWidgets))
				rwids := make(map[string]struct{}, len(readme.Widgets))
				upd := make(map[string]string)
				for _, nw := range newWidgets {
					for id := range nw {
						nwids[id] = struct{}{}
					}
				}
				for _, rw := range readme.Widgets {
					for id := range rw {
						rwids[id] = struct{}{}
					}
				}
				for nid := range nwids {
					if _, ex := rwids[nid]; !ex {
						upd["num_of_users"] = "+"
						if err := rs.WidgetRepo.Update(c, upd, nid); err != nil {
							return nil, err
						}
					}
				}
				for rid := range rwids {
					if _, ex := nwids[rid]; !ex {
						upd["num_of_users"] = "-"
						if err := rs.WidgetRepo.Update(c, upd, rid); err != nil {
							return nil, err
						}
					}
				}
			}
		}
		updates["last_update_time"] = now
		if err := rs.ReadmeRepo.Update(c, updates, id); err != nil {
			if fOk {
				if cerr := rs.CloudStorage.DeleteImage(c, newPid); cerr != nil {
					return nil, fmt.Errorf("%w : %w", err, cerr)
				}
			}
			return nil, err
		}
		if fOk {
			pId, err := rs.CloudStorage.GetPIdFromURL(oldURL)
			if err != nil {
				return nil, err
			}
			if err := rs.CloudStorage.DeleteImage(c, pId); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}); err != nil {
		log.Log.Error("faield to update readme")
		return errs.NewAppError(op, err)
	}

	log.Log.Info("readme updated successfully")
	return nil
}

func (rs *readmeServ) Get(ctx context.Context, id string) (*models.Readme, error) {
	op := "readmeServ.Get"
	log := rs.Logger.AddOp(op)
	log.Log.Info("receiving readme")
	readme, err := rs.ReadmeRepo.Get(ctx, id)
	if err != nil {
		log.Log.Error("failed to receive readme", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	log.Log.Info("readme received successfully")
	return readme, nil
}

func (rs *readmeServ) FetchByUser(ctx context.Context, amount, page uint, uid string) ([]dto.ReadmeResponse, error) {
	op := "readmeServ.FetchByUser"
	log := rs.Logger.AddOp(op)
	log.Log.Info("receiving readmes by user")
	rdms, err := rs.ReadmeRepo.FetchByUser(ctx, amount, page, uid)
	if err != nil {
		log.Log.Error("failed to receive readmes", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	readmes := make([]dto.ReadmeResponse, 0, len(rdms))
	for _, r := range rdms {
		readme := dto.ReadmeResponse{
			Id:             r.Id.String(),
			Title:          r.Title,
			Image:          r.Image,
			LastUpdateTime: r.LastUpdateTime,
			CreateTime:     r.CreateTime,
		}
		readmes = append(readmes, readme)
	}
	log.Log.Info("readmes received successfully")
	return readmes, nil
}
