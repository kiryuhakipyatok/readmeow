package cloudstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"readmeow/internal/config"
	"readmeow/pkg/errs"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudStorage interface {
	UploadImage(ctx context.Context, file io.Reader, filename, folder string) (string, string, error)
	DeleteImage(ctx context.Context, pid string) error
	GetPIdFromURL(url string) (string, error)
}

type cloudStorage struct {
	Cloud  *cloudinary.Cloudinary
	Config config.CloudStorageConfig
}

func MustConnect(cfg config.CloudStorageConfig) CloudStorage {
	cld, err := cloudinary.NewFromURL(cfg.CloudURL)
	if err != nil {
		panic(fmt.Errorf("failed to connect to cloudinary: %w", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if _, err := cld.Admin.Ping(ctx); err != nil {
		panic(fmt.Errorf("failed to ping cloudinary: %w", err))
	}

	return &cloudStorage{
		Cloud: cld,
	}
}

func (cs *cloudStorage) UploadImage(ctx context.Context, file io.Reader, filename string, folder string) (string, string, error) {
	op := "cloudStorage.UploadImage"
	ptr := func(b bool) *bool {
		return &b
	}

	res, err := cs.Cloud.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:     folder,
		PublicID:   filename,
		Overwrite:  ptr(true),
		Invalidate: ptr(true),
	})
	if err != nil {
		return "", "", errs.NewAppError(op, err)
	}

	pid := res.PublicID
	url := res.SecureURL

	return url, pid, nil
}

func (cs *cloudStorage) DeleteImage(ctx context.Context, pid string) error {
	op := "cloudStorage.DeleteImage"
	_, err := cs.Cloud.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: pid,
	})
	if err != nil {
		return errs.NewAppError(op, err)
	}
	return nil
}

func (cs *cloudStorage) GetPIdFromURL(url string) (string, error) {
	op := "cloudStorage.GetPIdFromURL"
	const prefix = "https://res.cloudinary.com/dt02alvlt/image/upload/"
	s, ok := strings.CutPrefix(url, prefix)
	if !ok {
		return "", errs.NewAppError(op, errors.New("prefix not found"))
	}
	parts := strings.SplitN(s, "/", 2)
	pid := strings.TrimSuffix(parts[1], filepath.Ext(s))
	if pid == "" {
		return "", errs.NewAppError(op, errors.New("failed to get pid from url"))
	}
	return pid, nil
}
