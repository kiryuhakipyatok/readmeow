package services

import (
	"context"
	"readmeow/internal/config"
	"readmeow/internal/domain/models"
	mock_repositories "readmeow/internal/domain/repositories/mocks"
	"readmeow/internal/dto"
	"readmeow/pkg/errs"
	"readmeow/pkg/logger"
	mock_storage "readmeow/pkg/storage/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWidgetServ_Get(t *testing.T) {
	type mockBehavior = func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string)
	nilUUID := uuid.Nil
	tests := []struct {
		name           string
		id             string
		expectedResult *models.Widget
		mockBehaviour  mockBehavior
		expectedError  error
	}{
		{
			name: "success",
			id:   nilUUID.String(),
			expectedResult: &models.Widget{
				Id:          nilUUID,
				Title:       "title",
				Image:       "image",
				Description: "description",
				Type:        "default",
				Tags:        map[string]any{},
				Likes:       10,
				Link:        "link",
				NumOfUsers:  12,
			},
			mockBehaviour: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string) {
				m.EXPECT().Get(ctx, id).Return(&models.Widget{
					Id:          nilUUID,
					Title:       "title",
					Image:       "image",
					Description: "description",
					Type:        "default",
					Tags:        map[string]any{},
					Likes:       10,
					Link:        "link",
					NumOfUsers:  12,
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:           "not found",
			id:             nilUUID.String(),
			expectedResult: nil,
			mockBehaviour: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string) {
				m.EXPECT().Get(ctx, id).Return(nil, errs.ErrNotFoundBase)
			},
			expectedError: errs.ErrNotFoundBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			userRepo := mock_repositories.NewMockUserRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			widgetService := NewWidgetServ(widgetRepo, userRepo, transactor, log)
			ctx := context.Background()
			tt.mockBehaviour(widgetRepo, ctx, tt.id)
			result, err := widgetService.Get(ctx, tt.id)
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

func TestWidgetServ_Search(t *testing.T) {
	nilUUID := uuid.Nil
	type mockBehavior = func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string)
	tests := []struct {
		name           string
		req            dto.SearchWidgetRequest
		mockBehavior   mockBehavior
		expectedResult []dto.WidgetResponse
		expectedError  error
	}{
		{
			name: "success without sort and filter with query",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				Query: "title",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, 0, len(sort))
					assert.Equal(t, 0, len(filter))
					return []models.Widget{
						{
							Id:          nilUUID,
							Title:       "title1",
							Image:       "image1",
							Description: "description1",
							Type:        "default1",
							Tags:        map[string]any{},
							Likes:       10,
							Link:        "link1",
							NumOfUsers:  12,
						},
						{
							Id:          nilUUID,
							Title:       "title2",
							Image:       "image2",
							Description: "description2",
							Type:        "default2",
							Tags:        map[string]any{},
							Likes:       12,
							Link:        "link2",
							NumOfUsers:  14,
						},
					}, nil
				})
			},
			expectedResult: []dto.WidgetResponse{
				{
					Id:          nilUUID.String(),
					Title:       "title1",
					Image:       "image1",
					Description: "description1",
					Likes:       10,
					NumOfUsers:  12,
				},
				{
					Id:          nilUUID.String(),
					Title:       "title2",
					Image:       "image2",
					Description: "description2",
					Likes:       12,
					NumOfUsers:  14,
				},
			},
			expectedError: nil,
		},
		{
			name: "success with sort (asc) and withot filter with query",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				SortWidgetsFields: dto.SortWidgetsFields{
					Likes:      "asc",
					NumOfUsers: "asc",
				},
				Query: "title",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, "asc", sort["Likes"])
					assert.Equal(t, "asc", sort["NumOfUsers"])
					assert.Equal(t, 0, len(filter))
					return []models.Widget{
						{
							Id:          nilUUID,
							Title:       "title1",
							Image:       "image1",
							Description: "description1",
							Type:        "default1",
							Tags:        map[string]any{},
							Likes:       10,
							Link:        "link1",
							NumOfUsers:  12,
						},
						{
							Id:          nilUUID,
							Title:       "title2",
							Image:       "image2",
							Description: "description2",
							Type:        "default2",
							Tags:        map[string]any{},
							Likes:       12,
							Link:        "link2",
							NumOfUsers:  14,
						},
					}, nil
				})
			},
			expectedResult: []dto.WidgetResponse{
				{
					Id:          nilUUID.String(),
					Title:       "title1",
					Image:       "image1",
					Description: "description1",
					Likes:       10,
					NumOfUsers:  12,
				},
				{
					Id:          nilUUID.String(),
					Title:       "title2",
					Image:       "image2",
					Description: "description2",
					Likes:       12,
					NumOfUsers:  14,
				},
			},
			expectedError: nil,
		},
		{
			name: "success with sort (desc) and withot filter with query",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				SortWidgetsFields: dto.SortWidgetsFields{
					Likes:      "desc",
					NumOfUsers: "desc",
				},
				Query: "title",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, "desc", sort["Likes"])
					assert.Equal(t, "desc", sort["NumOfUsers"])
					return []models.Widget{
						{
							Id:          nilUUID,
							Title:       "title2",
							Image:       "image2",
							Description: "description2",
							Type:        "default2",
							Tags:        map[string]any{},
							Likes:       12,
							Link:        "link2",
							NumOfUsers:  14,
						},
						{
							Id:          nilUUID,
							Title:       "title1",
							Image:       "image1",
							Description: "description1",
							Type:        "default1",
							Tags:        map[string]any{},
							Likes:       10,
							Link:        "link1",
							NumOfUsers:  12,
						},
					}, nil
				})
			},
			expectedResult: []dto.WidgetResponse{
				{
					Id:          nilUUID.String(),
					Title:       "title2",
					Image:       "image2",
					Description: "description2",
					Likes:       12,
					NumOfUsers:  14,
				},
				{
					Id:          nilUUID.String(),
					Title:       "title1",
					Image:       "image1",
					Description: "description1",
					Likes:       10,
					NumOfUsers:  12,
				},
			},
			expectedError: nil,
		},
		{
			name: "success without sort and with filter with query",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				FilterWidgetsFields: dto.FilterWidgetsFields{
					Tags:  []string{"daftag"},
					Types: []string{"default2"},
				},
				Query: "title",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.ElementsMatch(t, []string{"daftag"}, filter["Tags"])
					assert.ElementsMatch(t, []string{"default2"}, filter["Types"])
					assert.Equal(t, 0, len(sort))
					return []models.Widget{
						{
							Id:          nilUUID,
							Title:       "title2",
							Image:       "image2",
							Description: "description2",
							Type:        "default2",
							Tags: map[string]any{
								"deftag": "info",
							},
							Likes:      12,
							Link:       "link2",
							NumOfUsers: 14,
						},
					}, nil
				})
			},
			expectedResult: []dto.WidgetResponse{
				{
					Id:          nilUUID.String(),
					Title:       "title2",
					Image:       "image2",
					Description: "description2",
					Likes:       12,
					NumOfUsers:  14,
				},
			},
			expectedError: nil,
		},
		{
			name: "success with sort and filter with query",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				FilterWidgetsFields: dto.FilterWidgetsFields{
					Tags:  []string{"daftag"},
					Types: []string{"default2"},
				},
				SortWidgetsFields: dto.SortWidgetsFields{
					Likes:      "asc",
					NumOfUsers: "asc",
				},
				Query: "title",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, "asc", sort["Likes"])
					assert.Equal(t, "asc", sort["NumOfUsers"])
					assert.ElementsMatch(t, []string{"daftag"}, filter["Tags"])
					assert.ElementsMatch(t, []string{"default2"}, filter["Types"])
					return []models.Widget{
						{
							Id:          nilUUID,
							Title:       "title2",
							Image:       "image2",
							Description: "description2",
							Type:        "default2",
							Tags: map[string]any{
								"deftag": "info",
							},
							Likes:      12,
							Link:       "link2",
							NumOfUsers: 14,
						},
					}, nil
				})
			},
			expectedResult: []dto.WidgetResponse{
				{
					Id:          nilUUID.String(),
					Title:       "title2",
					Image:       "image2",
					Description: "description2",
					Likes:       12,
					NumOfUsers:  14,
				},
			},
			expectedError: nil,
		},
		{
			name: "empty success without sort and filter with query",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				Query: "widget",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, 0, len(sort))
					assert.Equal(t, 0, len(filter))
					return []models.Widget{}, nil
				})
			},
			expectedResult: []dto.WidgetResponse{},
			expectedError:  nil,
		},
		{
			name: "ivalid likes value in sort",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				SortWidgetsFields: dto.SortWidgetsFields{
					Likes: "qwe",
				},
				Query: "widget",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, 0, len(filter))
					assert.Equal(t, "qwe", sort["Likes"])
					return nil, errs.ErrInvalidValuesBase
				})
			},
			expectedResult: nil,
			expectedError:  errs.ErrInvalidValuesBase,
		},
		{
			name: "ivalid num_of_users value in sort",
			req: dto.SearchWidgetRequest{
				PaginationRequest: dto.PaginationRequest{
					Amount: 2,
					Page:   1,
				},
				SortWidgetsFields: dto.SortWidgetsFields{
					NumOfUsers: "acc",
				},
				Query: "widget",
			},
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, amount, page uint, query string) {
				m.EXPECT().Search(ctx, amount, page, query, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amount, page uint, query string, filter map[string][]string, sort map[string]string) ([]models.Widget, error) {
					assert.Equal(t, 0, len(filter))
					assert.Equal(t, "acc", sort["NumOfUsers"])
					return nil, errs.ErrInvalidValuesBase
				})
			},
			expectedResult: nil,
			expectedError:  errs.ErrInvalidValuesBase,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			userRepo := mock_repositories.NewMockUserRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			widgetService := NewWidgetServ(widgetRepo, userRepo, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(widgetRepo, ctx, tt.req.Amount, tt.req.Page, tt.req.Query)
			results, err := widgetService.Search(ctx, tt.req)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedResult, results)
			}
		})
	}
}

func TestWidgetServ_Like(t *testing.T) {
	nilUUID := uuid.Nil
	type mockBehavior = func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string)
	tests := []struct {
		name           string
		userId         string
		widgetId       string
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name:     "success",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Like(ctx, uid, id).Return(nil)
				mr.EXPECT().Update(ctx, gomock.Any(), id).DoAndReturn(func(ctx context.Context, updates map[string]string, id string) error {
					assert.Equal(t, "+", updates["likes"])
					return nil
				})
			},
			expectedResult: nil,
		},
		{
			name:     "user not found",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Like(ctx, uid, id).Return(errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name:     "widget not found",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Like(ctx, uid, id).Return(errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name:     "already exists",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Like(ctx, uid, id).Return(errs.ErrAlreadyExistsBase)
			},
			expectedResult: errs.ErrAlreadyExistsBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			userRepo := mock_repositories.NewMockUserRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			widgetService := NewWidgetServ(widgetRepo, userRepo, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(widgetRepo, transactor, ctx, tt.userId, tt.widgetId)
			res := widgetService.Like(ctx, tt.widgetId, tt.userId)
			if tt.expectedResult != nil {
				assert.Error(t, res)
				assert.ErrorIs(t, res, tt.expectedResult)
			} else {
				assert.NoError(t, res)
			}
		})
	}
}

func TestWidgetServ_Dislike(t *testing.T) {
	nilUUID := uuid.Nil
	type mockBehavior = func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string)
	tests := []struct {
		name           string
		userId         string
		widgetId       string
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name:     "success",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Dislike(ctx, uid, id).Return(nil)
				mr.EXPECT().Update(ctx, gomock.Any(), id).DoAndReturn(func(ctx context.Context, updates map[string]string, id string) error {
					assert.Equal(t, "-", updates["likes"])
					return nil
				})
			},
			expectedResult: nil,
		},
		{
			name:     "user not found",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Dislike(ctx, uid, id).Return(errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name:     "widget not found",
			userId:   nilUUID.String(),
			widgetId: nilUUID.String(),
			mockBehavior: func(mr *mock_repositories.MockWidgetRepo, mt *mock_storage.MockTransactor, ctx context.Context, id, uid string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				mr.EXPECT().Dislike(ctx, uid, id).Return(errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			userRepo := mock_repositories.NewMockUserRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			widgetService := NewWidgetServ(widgetRepo, userRepo, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(widgetRepo, transactor, ctx, tt.userId, tt.widgetId)
			res := widgetService.Dislike(ctx, tt.widgetId, tt.userId)
			if tt.expectedResult != nil {
				assert.Error(t, res)
				assert.ErrorIs(t, res, tt.expectedResult)
			} else {
				assert.NoError(t, res)
			}
		})
	}
}

func TestWidgetServ_FetchFavorite(t *testing.T) {
	nilUUID := uuid.Nil
	type mockBehavior = func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string, amount, page uint)
	tests := []struct {
		name           string
		id             string
		amount, page   uint
		mockBehavior   mockBehavior
		expectedResult []dto.WidgetResponse
		expectedError  error
	}{
		{
			name:   "success",
			id:     nilUUID.String(),
			amount: 2,
			page:   1,
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string, amount, page uint) {
				m.EXPECT().FetchFavorite(ctx, id, amount, page).Return(
					[]models.Widget{
						{
							Id:          nilUUID,
							Title:       "title2",
							Image:       "image2",
							Description: "description2",
							Type:        "default2",
							Tags:        map[string]any{},
							Likes:       12,
							Link:        "link2",
							NumOfUsers:  14,
						},
						{
							Id:          nilUUID,
							Title:       "title1",
							Image:       "image1",
							Description: "description1",
							Type:        "default1",
							Tags:        map[string]any{},
							Likes:       10,
							Link:        "link1",
							NumOfUsers:  12,
						},
					}, nil)
			},
			expectedResult: []dto.WidgetResponse{
				{
					Id:          nilUUID.String(),
					Title:       "title2",
					Image:       "image2",
					Description: "description2",
					Likes:       12,
					NumOfUsers:  14,
				},
				{
					Id:          nilUUID.String(),
					Title:       "title1",
					Image:       "image1",
					Description: "description1",
					Likes:       10,
					NumOfUsers:  12,
				},
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			id:     nilUUID.String(),
			amount: 2,
			page:   1,
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string, amount, page uint) {
				m.EXPECT().FetchFavorite(ctx, id, amount, page).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: nil,
			expectedError:  errs.ErrNotFoundBase,
		},
		{
			name:   "empty success",
			id:     nilUUID.String(),
			amount: 2,
			page:   1,
			mockBehavior: func(m *mock_repositories.MockWidgetRepo, ctx context.Context, id string, amount, page uint) {
				m.EXPECT().FetchFavorite(ctx, id, amount, page).Return([]models.Widget{}, nil)
			},
			expectedResult: []dto.WidgetResponse{},
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			widgetRepo := mock_repositories.NewMockWidgetRepo(c)
			userRepo := mock_repositories.NewMockUserRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			widgetService := NewWidgetServ(widgetRepo, userRepo, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(widgetRepo, ctx, tt.id, tt.amount, tt.page)
			results, err := widgetService.FetchFavorite(ctx, tt.id, tt.amount, tt.page)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedResult, results)
			}
		})
	}
}
