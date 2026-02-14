package services

import (
	"context"
	"readmeow/internal/config"
	"readmeow/internal/domain/models"
	mock_repositories "readmeow/internal/domain/repositories/mocks"
	"readmeow/internal/domain/services/utils"
	"readmeow/internal/dto"
	mock_cloudstorage "readmeow/pkg/cloudstorage/mocks"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	mock_storage "readmeow/pkg/storage/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestReadmeServ_Get(t *testing.T) {
	type mockBehavior = func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, id string)
	nilUUID := uuid.Nil
	time := time.Now()
	tests := []struct {
		name           string
		id             string
		mockBehavior   mockBehavior
		expectedResult *models.Readme
		expectedError  error
	}{
		{
			name: "success",
			id:   nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, id string) {
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{
					Id:             nilUUID,
					OwnerId:        nilUUID,
					TemplateId:     nilUUID,
					Title:          "title",
					Image:          "image",
					Text:           []string{},
					Links:          []string{},
					Widgets:        []map[string]string{},
					RenderOrder:    []string{},
					CreateTime:     time,
					LastUpdateTime: time,
				}, nil)
			},
			expectedResult: &models.Readme{
				Id:             nilUUID,
				OwnerId:        nilUUID,
				TemplateId:     nilUUID,
				Title:          "title",
				Image:          "image",
				Text:           []string{},
				Links:          []string{},
				Widgets:        []map[string]string{},
				RenderOrder:    []string{},
				CreateTime:     time,
				LastUpdateTime: time,
			},
			expectedError: nil,
		},

		{
			name: "not found",
			id:   nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, id string) {
				mr.EXPECT().Get(ctx, id).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: nil,
			expectedError:  errs.ErrNotFoundBase,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			readmeRepo := mock_repositories.NewMockReadmeRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			readmeServ := NewReadmeServ(readmeRepo, userRepo, templRepo, widgetRepo, transactor, clouds, log)
			ctx := context.Background()
			tt.mockBehavior(readmeRepo, ctx, tt.id)
			result, err := readmeServ.Get(ctx, tt.id)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestReadmeServ_Delete(t *testing.T) {
	type mockBehavior = func(mr *mock_repositories.MockReadmeRepo, mu *mock_repositories.MockUserRepo, mw *mock_repositories.MockWidgetRepo, mt *mock_repositories.MockTemplateRepo, mc *mock_cloudstorage.MockCloudStorage, mtr *mock_storage.MockTransactor, ctx context.Context, id, uid string)
	nilUUID := uuid.Nil
	uuid := uuid.New()
	testWidgets := []map[string]string{
		{
			"wid1": "1",
		},
		{
			"wid2": "2",
		},
	}
	tests := []struct {
		name           string
		id, uid        string
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name: "success",
			id:   nilUUID.String(),
			uid:  nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mu *mock_repositories.MockUserRepo, mw *mock_repositories.MockWidgetRepo, mt *mock_repositories.MockTemplateRepo, mc *mock_cloudstorage.MockCloudStorage, mtr *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, TemplateId: nilUUID, Widgets: testWidgets, OwnerId: nilUUID}, nil)
				for range len(testWidgets) {
					mw.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updates map[string]string, _ string) error {
						assert.Equal(t, "-", updates["num_of_users"])
						return nil
					})
				}

				mt.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updates map[string]any, _ string) error {
					assert.Equal(t, "-", updates["num_of_users"])
					return nil
				})
				mr.EXPECT().Delete(ctx, id).Return(nil)
				mc.EXPECT().GetPIdFromURL(gomock.Any()).Return("pid", nil)
				mc.EXPECT().DeleteImage(ctx, "pid").Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "user not found",
			id:   nilUUID.String(),
			uid:  nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mu *mock_repositories.MockUserRepo, mw *mock_repositories.MockWidgetRepo, mt *mock_repositories.MockTemplateRepo, mc *mock_cloudstorage.MockCloudStorage, mtr *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name: "readme not found",
			id:   nilUUID.String(),
			uid:  nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mu *mock_repositories.MockUserRepo, mw *mock_repositories.MockWidgetRepo, mt *mock_repositories.MockTemplateRepo, mc *mock_cloudstorage.MockCloudStorage, mtr *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mr.EXPECT().Get(ctx, id).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name: "user is not a owner",
			id:   nilUUID.String(),
			uid:  nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mu *mock_repositories.MockUserRepo, mw *mock_repositories.MockWidgetRepo, mt *mock_repositories.MockTemplateRepo, mc *mock_cloudstorage.MockCloudStorage, mtr *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, TemplateId: nilUUID, Widgets: testWidgets, OwnerId: uuid}, nil)
			},
			expectedResult: errs.ErrUserIsNotAOwnerBase,
		},
		{
			name: "success with readme without widgets",
			id:   nilUUID.String(),
			uid:  nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mu *mock_repositories.MockUserRepo, mw *mock_repositories.MockWidgetRepo, mt *mock_repositories.MockTemplateRepo, mc *mock_cloudstorage.MockCloudStorage, mtr *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, TemplateId: nilUUID, Widgets: nil, OwnerId: nilUUID}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updates map[string]any, _ string) error {
					assert.Equal(t, "-", updates["num_of_users"])
					return nil
				})
				mr.EXPECT().Delete(ctx, id).Return(nil)
				mc.EXPECT().GetPIdFromURL(gomock.Any()).Return("pid", nil)
				mc.EXPECT().DeleteImage(ctx, "pid").Return(nil)
			},
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			readmeRepo := mock_repositories.NewMockReadmeRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			readmeServ := NewReadmeServ(readmeRepo, userRepo, templRepo, widgetRepo, transactor, clouds, log)
			ctx := context.Background()
			tt.mockBehavior(readmeRepo, userRepo, widgetRepo, templRepo, clouds, transactor, ctx, tt.id, tt.uid)
			res := readmeServ.Delete(ctx, tt.id, tt.uid)
			if tt.expectedResult != nil {
				assert.Error(t, res)
				assert.ErrorIs(t, res, tt.expectedResult)
			} else {
				assert.NoError(t, res)
				assert.Equal(t, tt.expectedResult, res)
			}
		})
	}
}

func TestReadmeServ_FetchByUser(t *testing.T) {
	type mockBehavior = func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, amount, page uint, uid string)
	nilUUID := uuid.Nil
	tests := []struct {
		name           string
		amount, page   uint
		uid            string
		mockBehavior   mockBehavior
		expectedResult []dto.ReadmeResponse
		expectedError  error
	}{
		{
			name:   "success",
			amount: 1,
			page:   2,
			uid:    nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, amount, page uint, uid string) {
				uidUUID, _ := uuid.Parse(uid)
				mr.EXPECT().FetchByUser(ctx, amount, page, uid).Return([]models.Readme{
					{
						Id:      nilUUID,
						OwnerId: uidUUID,
						Title:   "t1",
					},
					{
						Id:      nilUUID,
						OwnerId: uidUUID,
						Title:   "t2",
					},
				}, nil)
			},
			expectedResult: []dto.ReadmeResponse{
				{
					Id:    nilUUID.String(),
					Title: "t1",
				},
				{
					Id:    nilUUID.String(),
					Title: "t2",
				},
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			amount: 1,
			page:   2,
			uid:    nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, amount, page uint, uid string) {
				mr.EXPECT().FetchByUser(ctx, amount, page, uid).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: nil,
			expectedError:  errs.ErrNotFoundBase,
		},
		{
			name:   "success without readmes",
			amount: 1,
			page:   2,
			uid:    nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, ctx context.Context, amount, page uint, uid string) {
				mr.EXPECT().FetchByUser(ctx, amount, page, uid).Return([]models.Readme{}, nil)
			},
			expectedResult: []dto.ReadmeResponse{},
			expectedError:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			readmeRepo := mock_repositories.NewMockReadmeRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			readmeServ := NewReadmeServ(readmeRepo, userRepo, templRepo, widgetRepo, transactor, clouds, log)
			ctx := context.Background()
			tt.mockBehavior(readmeRepo, ctx, tt.amount, tt.page, tt.uid)
			res, err := readmeServ.FetchByUser(ctx, tt.amount, tt.page, tt.uid)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedResult, res)
			}
		})
	}
}

func TestReadmeServ_Update(t *testing.T) {
	type mockBehavior = func(mr *mock_repositories.MockReadmeRepo, mt *mock_storage.MockTransactor, mc *mock_cloudstorage.MockCloudStorage, mw *mock_repositories.MockWidgetRepo, ctx context.Context, updates map[string]any, id string)
	nilUUID := uuid.Nil
	time := time.Now()
	tests := []struct {
		name           string
		id             string
		updates        map[string]any
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name: "success without image (widget additing)",
			id:   nilUUID.String(),
			updates: map[string]any{
				"title":            "title",
				"text":             map[string]string{},
				"links":            []string{},
				"widgets":          []map[string]string{{nilUUID.String(): "w"}},
				"render_order":     []string{},
				"last_update_time": time,
			},
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mt *mock_storage.MockTransactor, mc *mock_cloudstorage.MockCloudStorage, mw *mock_repositories.MockWidgetRepo, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, Widgets: []map[string]string{}}, nil)
				for range len(updates["widgets"].([]map[string]string)) {
					mw.EXPECT().Update(ctx, map[string]string{"num_of_users": "+"}, nilUUID.String()).Return(nil)
				}
				mr.EXPECT().Update(ctx, updates, id).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "success without image (widget deleting)",
			id:   nilUUID.String(),
			updates: map[string]any{
				"title":            "title",
				"text":             map[string]string{},
				"links":            []string{},
				"widgets":          []map[string]string{{nilUUID.String(): "w"}},
				"render_order":     []string{},
				"last_update_time": time,
			},
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mt *mock_storage.MockTransactor, mc *mock_cloudstorage.MockCloudStorage, mw *mock_repositories.MockWidgetRepo, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, Widgets: []map[string]string{{nilUUID.String(): "w"}, {nilUUID.String() + "1": "q"}}}, nil)
				for range len(updates["widgets"].([]map[string]string)) {
					mw.EXPECT().Update(ctx, map[string]string{"num_of_users": "-"}, nilUUID.String()+"1").Return(nil)
				}
				mr.EXPECT().Update(ctx, updates, id).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "success with image",
			id:   nilUUID.String(),
			updates: map[string]any{
				"title":            "title",
				"text":             map[string]string{},
				"links":            []string{},
				"widgets":          []map[string]string{{nilUUID.String(): "w"}},
				"render_order":     []string{},
				"last_update_time": time,
				"image":            utils.NewTestFileHeader(),
			},
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mt *mock_storage.MockTransactor, mc *mock_cloudstorage.MockCloudStorage, mw *mock_repositories.MockWidgetRepo, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mc.EXPECT().UploadImage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return("url", "newpid", nil)
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, Widgets: []map[string]string{}}, nil)
				for range len(updates["widgets"].([]map[string]string)) {
					mw.EXPECT().Update(ctx, map[string]string{"num_of_users": "+"}, nilUUID.String()).Return(nil)
				}
				mr.EXPECT().Update(ctx, updates, id).Return(nil)
				mc.EXPECT().GetPIdFromURL(gomock.Any()).Return("oldpid", nil)
				mc.EXPECT().DeleteImage(ctx, "oldpid").Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "not found",
			id:   nilUUID.String(),
			updates: map[string]any{
				"title":            "title",
				"text":             map[string]string{},
				"links":            []string{},
				"widgets":          []map[string]string{{nilUUID.String(): "w"}},
				"render_order":     []string{},
				"last_update_time": time,
			},
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mt *mock_storage.MockTransactor, mc *mock_cloudstorage.MockCloudStorage, mw *mock_repositories.MockWidgetRepo, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Get(ctx, id).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name: "invalid fields",
			id:   nilUUID.String(),
			updates: map[string]any{
				"title":        "title",
				"text":         map[string]string{},
				"links":        []string{},
				"widgets":      []map[string]string{{nilUUID.String(): "w"}},
				"render_order": []string{},
				"qwd":          "qwd",
			},
			mockBehavior: func(mr *mock_repositories.MockReadmeRepo, mt *mock_storage.MockTransactor, mc *mock_cloudstorage.MockCloudStorage, mw *mock_repositories.MockWidgetRepo, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Get(ctx, id).Return(&models.Readme{Id: nilUUID, Widgets: []map[string]string{}}, nil)
				for range len(updates["widgets"].([]map[string]string)) {
					mw.EXPECT().Update(ctx, map[string]string{"num_of_users": "+"}, nilUUID.String()).Return(nil)
				}
				mr.EXPECT().Update(ctx, updates, id).Return(errs.ErrInvalidFieldsBase)
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			readmeRepo := mock_repositories.NewMockReadmeRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			readmeServ := NewReadmeServ(readmeRepo, userRepo, templRepo, widgetRepo, transactor, clouds, log)
			ctx := context.Background()
			tt.mockBehavior(readmeRepo, transactor, clouds, widgetRepo, ctx, tt.updates, tt.id)
			res := readmeServ.Update(ctx, tt.updates, tt.id)
			if tt.expectedResult != nil {
				assert.Error(t, res)
				assert.ErrorIs(t, res, tt.expectedResult)
			} else {
				assert.NoError(t, res)
				assert.Equal(t, tt.expectedResult, res)
			}
		})
	}
}

func TestReadmeServ_Create(t *testing.T) {
	nilUUID := uuid.Nil
	type mockBehavior = func(
		mr *mock_repositories.MockReadmeRepo,
		mtr *mock_storage.MockTransactor,
		mt *mock_repositories.MockTemplateRepo,
		mu *mock_repositories.MockUserRepo,
		mw *mock_repositories.MockWidgetRepo,
		mc *mock_cloudstorage.MockCloudStorage,
		ctx context.Context,
		uid string,
		req *dto.CreateReadmeRequest,
	)
	tests := []struct {
		name           string
		uid            string
		req            *dto.CreateReadmeRequest
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name: "success with tempalte",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, req.TemplateId).Return(&models.TemplateWithOwner{Template: models.Template{Id: nilUUID}}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mw.EXPECT().GetByIds(ctx, gomock.Any()).Return([]models.Widget{{Id: nilUUID}}, nil)
				for range req.Widgets {
					mw.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				}
				mu.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mc.EXPECT().UploadImage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return("url", "pid", nil)
				mr.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "user not found",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name: "template not found",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, req.TemplateId).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name: "already exists",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, req.TemplateId).Return(&models.TemplateWithOwner{Template: models.Template{Id: nilUUID}}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mw.EXPECT().GetByIds(ctx, gomock.Any()).Return([]models.Widget{{Id: nilUUID}}, nil)
				for range req.Widgets {
					mw.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				}
				mu.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mc.EXPECT().UploadImage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return("url", "pid", nil)
				mr.EXPECT().Create(ctx, gomock.Any()).Return(errs.ErrAlreadyExistsBase)
				mc.EXPECT().DeleteImage(ctx, "pid").Return(nil)
			},
			expectedResult: errs.ErrAlreadyExistsBase,
		},
		{
			name: "widget invalid fields",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, req.TemplateId).Return(&models.TemplateWithOwner{Template: models.Template{Id: nilUUID}}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mw.EXPECT().GetByIds(ctx, gomock.Any()).Return([]models.Widget{{Id: nilUUID}}, nil)
				for range req.Widgets {
					mw.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(errs.ErrInvalidFieldsBase)
				}
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
		{
			name: "template invalid fields",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, req.TemplateId).Return(&models.TemplateWithOwner{Template: models.Template{Id: nilUUID}}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mw.EXPECT().GetByIds(ctx, gomock.Any()).Return([]models.Widget{{Id: nilUUID}}, nil)
				for range req.Widgets {
					mw.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				}
				mu.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(errs.ErrInvalidFieldsBase)
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
		{
			name: "user invalid fields",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				TemplateId:  nilUUID.String(),
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, req.TemplateId).Return(&models.TemplateWithOwner{Template: models.Template{Id: nilUUID}}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(errs.ErrInvalidFieldsBase)
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
		{
			name: "success without tempalte",
			uid:  nilUUID.String(),
			req: &dto.CreateReadmeRequest{
				Image:       utils.NewTestFileHeader(),
				Title:       "title",
				RenderOrder: []string{"q", "w"},
				Text:        []string{"wqd", "wqd"},
				Links:       []string{"qwd"},
				Widgets:     []map[string]string{{nilUUID.String(): "wqd"}},
			},
			mockBehavior: func(
				mr *mock_repositories.MockReadmeRepo,
				mtr *mock_storage.MockTransactor,
				mt *mock_repositories.MockTemplateRepo,
				mu *mock_repositories.MockUserRepo,
				mw *mock_repositories.MockWidgetRepo,
				mc *mock_cloudstorage.MockCloudStorage,
				ctx context.Context,
				uid string,
				req *dto.CreateReadmeRequest,
			) {
				mtr.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mu.EXPECT().Get(ctx, uid).Return(&models.User{Id: nilUUID}, nil)
				mt.EXPECT().Get(ctx, nilUUID.String()).Return(&models.TemplateWithOwner{Template: models.Template{Id: nilUUID}}, nil)
				mt.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mw.EXPECT().GetByIds(ctx, gomock.Any()).Return([]models.Widget{{Id: nilUUID}}, nil)
				for range req.Widgets {
					mw.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				}
				mu.EXPECT().Update(ctx, gomock.Any(), nilUUID.String()).Return(nil)
				mc.EXPECT().UploadImage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return("url", "pid", nil)
				mr.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			},
			expectedResult: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			readmeRepo := mock_repositories.NewMockReadmeRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			readmeServ := NewReadmeServ(readmeRepo, userRepo, templRepo, widgetRepo, transactor, clouds, log)
			ctx := context.Background()
			tt.mockBehavior(readmeRepo, transactor, templRepo, userRepo, widgetRepo, clouds, ctx, tt.uid, tt.req)
			res := readmeServ.Create(ctx, tt.uid, tt.req)
			if tt.expectedResult != nil {
				assert.Error(t, res)
				assert.ErrorIs(t, res, tt.expectedResult)
			} else {
				assert.NoError(t, res)
				assert.Equal(t, tt.expectedResult, res)
			}
		})
	}
}
