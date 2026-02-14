package services

import (
	"context"
	"crypto/sha256"
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
	"golang.org/x/crypto/bcrypt"
)

func TestUsersServ_Get(t *testing.T) {
	nilUUID := uuid.Nil
	now := time.Now()
	type mockBehavior = func(um *mock_repositories.MockUserRepo, tm *mock_repositories.MockTemplateRepo, ctx context.Context, id string, showPrivate bool)
	tests := []struct {
		name           string
		id             string
		showPrivate    bool
		mockBehavior   mockBehavior
		expectedResult *dto.UserResponse
		expectedError  error
	}{
		{
			name:        "success for profile",
			id:          nilUUID.String(),
			showPrivate: true,
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_repositories.MockTemplateRepo, ctx context.Context, id string, showPrivate bool) {
				um.EXPECT().Get(ctx, id).Return(&models.User{
					Id: nilUUID,
					Credentials: models.Credentials{
						Nickname:   "nick",
						Email:      "email",
						Login:      utils.StrToPtr("login"),
						Password:   []byte{},
						Provider:   "local",
						ProviderId: nil,
					},
					NumOfReadmes:   0,
					NumOfTemplates: 0,
					TimeOfRegister: now,
					Avatar:         "ava.png",
				}, nil)
				tm.EXPECT().FetchByUser(ctx, id, showPrivate)
			},
			expectedResult: &dto.UserResponse{
				Id:             nilUUID.String(),
				Nickname:       "nick",
				Email:          "email",
				Avatar:         "ava.png",
				NumOfReadmes:   0,
				NumOfTemplates: 0,
				TimeOfRegister: now,
				Templates:      []dto.TemplateInfo{},
			},
			expectedError: nil,
		},
		{
			name:        "success for user",
			id:          nilUUID.String(),
			showPrivate: false,
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_repositories.MockTemplateRepo, ctx context.Context, id string, showPrivate bool) {
				um.EXPECT().Get(ctx, id).Return(&models.User{
					Id: nilUUID,
					Credentials: models.Credentials{
						Nickname:   "nick",
						Email:      "email",
						Login:      utils.StrToPtr("login"),
						Password:   []byte{},
						Provider:   "local",
						ProviderId: nil,
					},
					NumOfReadmes:   0,
					NumOfTemplates: 0,
					TimeOfRegister: now,
					Avatar:         "ava.png",
				}, nil)
				tm.EXPECT().FetchByUser(ctx, id, showPrivate)
			},
			expectedResult: &dto.UserResponse{
				Id:             nilUUID.String(),
				Nickname:       "nick",
				Email:          "email",
				Avatar:         "ava.png",
				NumOfReadmes:   0,
				NumOfTemplates: 0,
				TimeOfRegister: time.Now(),
				Templates:      []dto.TemplateInfo{},
			},
			expectedError: nil,
		},
		{
			name:        "not found",
			id:          nilUUID.String(),
			showPrivate: false,
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_repositories.MockTemplateRepo, ctx context.Context, id string, showPrivate bool) {
				um.EXPECT().Get(ctx, id).Return(nil, errs.ErrNotFoundBase)
				tm.EXPECT().FetchByUser(ctx, id, showPrivate).Return(nil, errs.ErrNotFoundBase)
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
			passwRepo := mock_repositories.NewMockPasswordResetRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			userServ := NewUserServ(userRepo, passwRepo, templRepo, clouds, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(userRepo, templRepo, ctx, tt.id, tt.showPrivate)
			res, err := userServ.Get(ctx, tt.id, tt.showPrivate)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, res)
			}
		})
	}
}

func TestUsersServ_Update(t *testing.T) {
	nilUUID := uuid.Nil
	type mockBehavior = func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string)
	tests := []struct {
		name           string
		updates        map[string]any
		id             string
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name: "success with num_of_readmes",
			updates: map[string]any{
				"num_of_readmes": "+",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name:    "not found",
			updates: map[string]any{},
			id:      nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name: "success with nickname",
			updates: map[string]any{
				"nickname": "nick",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "success with num_of_templates",
			updates: map[string]any{
				"num_of_templates": "-",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "success with avatar",
			updates: map[string]any{
				"avatar": utils.NewTestFileHeader(),
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetAvatar(ctx, id).Return("oldurl", nil)
				cs.EXPECT().UploadImage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return("someurl", "newpid", nil)
				um.EXPECT().Update(ctx, updates, id).DoAndReturn(func(ctx context.Context, updates map[string]any, id string) error {
					assert.Equal(t, "someurl", updates["avatar"])
					return nil
				})
				cs.EXPECT().GetPIdFromURL("oldurl").Return("oldpid", nil)
				cs.EXPECT().DeleteImage(ctx, "oldpid").Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "success with avatar and num_of_readmes",
			updates: map[string]any{
				"avatar":         utils.NewTestFileHeader(),
				"num_of_readmes": "+",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetAvatar(ctx, id).Return("oldurl", nil)
				cs.EXPECT().UploadImage(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return("someurl", "newpid", nil)
				um.EXPECT().Update(ctx, updates, id).DoAndReturn(func(ctx context.Context, updates map[string]any, id string) error {
					assert.Equal(t, "someurl", updates["avatar"])
					return nil
				})
				cs.EXPECT().GetPIdFromURL("oldurl").Return("oldpid", nil)
				cs.EXPECT().DeleteImage(ctx, "oldpid").Return(nil)
			},
			expectedResult: nil,
		},
		{
			name: "invalid fieds",
			updates: map[string]any{
				"qwe": "qwe",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(errs.ErrInvalidFieldsBase)
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
		{
			name: "invalid fieds for num_of_readmes",
			updates: map[string]any{
				"num_of_readmes": "qwq",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(errs.ErrInvalidFieldsBase)
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
		{
			name: "invalid fieds for num_of_templates",
			updates: map[string]any{
				"templates": "qwq",
			},
			id: nilUUID.String(),
			mockBehavior: func(um *mock_repositories.MockUserRepo, mt *mock_storage.MockTransactor, cs *mock_cloudstorage.MockCloudStorage, ctx context.Context, updates map[string]any, id string) {
				mt.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().Update(ctx, updates, id).Return(errs.ErrInvalidFieldsBase)
			},
			expectedResult: errs.ErrInvalidFieldsBase,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			passwRepo := mock_repositories.NewMockPasswordResetRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			userServ := NewUserServ(userRepo, passwRepo, templRepo, clouds, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(userRepo, transactor, clouds, ctx, tt.updates, tt.id)
			res := userServ.Update(ctx, tt.updates, tt.id)
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

func TestUsersServ_Delete(t *testing.T) {
	type mockBehavior = func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, password string)
	nilUUID := uuid.Nil.String()
	testPass, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	tests := []struct {
		name           string
		id             string
		password       string
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name:     "success",
			id:       nilUUID,
			password: "password",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, password string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetPassword(ctx, id).Return(testPass, nil)
				um.EXPECT().Delete(ctx, id).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name:     "not found",
			id:       nilUUID,
			password: "password",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, password string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetPassword(ctx, id).Return(nil, errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name:     "incorrect password",
			id:       nilUUID,
			password: "password2",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, password string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetPassword(ctx, id).Return(nil, errs.ErrIncorrectOldPasswordBase)
			},
			expectedResult: errs.ErrIncorrectOldPasswordBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			passwRepo := mock_repositories.NewMockPasswordResetRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			userServ := NewUserServ(userRepo, passwRepo, templRepo, clouds, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(userRepo, transactor, ctx, tt.id, tt.password)
			res := userServ.Delete(ctx, tt.id, tt.password)
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

func TestUsersServ_ChangePassword(t *testing.T) {
	type mockBehavior = func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, oldPassword, newPasswrod string)
	nilUUID := uuid.Nil.String()
	tests := []struct {
		name           string
		id             string
		oldPassword    string
		mockBehavior   mockBehavior
		newPassword    string
		expectedResult error
	}{
		{
			name:        "success",
			id:          nilUUID,
			newPassword: "newpassword",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, oldPassword, newPassword string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				oldHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), 12)
				um.EXPECT().GetPassword(ctx, id).Return(oldHash, nil)
				um.EXPECT().ChangePassword(ctx, id, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, newHash []byte) error {
					bcrypt.CompareHashAndPassword(newHash, []byte(newPassword))
					return nil
				})
			},
			oldPassword:    "oldpassword",
			expectedResult: nil,
		},
		{
			name:        "not found",
			id:          nilUUID,
			newPassword: "newpassword",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, oldPassword, newPassword string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetPassword(ctx, id).Return(nil, errs.ErrNotFoundBase)
			},
			oldPassword:    "oldpassword",
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name:        "incorrect old password",
			id:          nilUUID,
			newPassword: "newpassword",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, ctx context.Context, id, oldPassword, newPassword string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetPassword(ctx, id).Return(nil, errs.ErrIncorrectOldPasswordBase)
			},
			oldPassword:    "oldpassword",
			expectedResult: errs.ErrIncorrectOldPasswordBase,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			passwRepo := mock_repositories.NewMockPasswordResetRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			userServ := NewUserServ(userRepo, passwRepo, templRepo, clouds, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(userRepo, transactor, ctx, tt.id, tt.oldPassword, tt.newPassword)
			res := userServ.ChangePassword(ctx, tt.id, tt.oldPassword, tt.newPassword)
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

func TestUsersServ_ResetPassword(t *testing.T) {
	type mockBehavior = func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, pm *mock_repositories.MockPasswordResetRepo, ctx context.Context, email, token, newPassword string)
	nilUUID := uuid.Nil.String()
	tests := []struct {
		name           string
		email          string
		token          string
		newPassword    string
		mockBehavior   mockBehavior
		expectedResult error
	}{
		{
			name:        "success",
			email:       "email",
			token:       nilUUID,
			newPassword: "newpassword",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, pm *mock_repositories.MockPasswordResetRepo, ctx context.Context, email, token, newPassword string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetByEmail(ctx, email).Return(nilUUID, nil)
				tokenHash := sha256.Sum256([]byte(token))
				pm.EXPECT().CheckToken(ctx, nilUUID, tokenHash[:]).Return(true, nil)
				um.EXPECT().ChangePassword(ctx, nilUUID, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, newHash []byte) error {
					bcrypt.CompareHashAndPassword(newHash, []byte(newPassword))
					return nil
				})
				pm.EXPECT().Delete(ctx, nilUUID).Return(nil)
			},
			expectedResult: nil,
		},
		{
			name:        "not found",
			email:       "email",
			token:       nilUUID,
			newPassword: "newpassword",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, pm *mock_repositories.MockPasswordResetRepo, ctx context.Context, email, token, newPassword string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetByEmail(ctx, email).Return("", errs.ErrNotFoundBase)
			},
			expectedResult: errs.ErrNotFoundBase,
		},
		{
			name:        "invalid token",
			email:       "email",
			token:       nilUUID,
			newPassword: "newpassword",
			mockBehavior: func(um *mock_repositories.MockUserRepo, tm *mock_storage.MockTransactor, pm *mock_repositories.MockPasswordResetRepo, ctx context.Context, email, token, newPassword string) {
				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
					_, err := f(txCtx)
					return nil, err
				})
				um.EXPECT().GetByEmail(ctx, email).Return(nilUUID, nil)
				tokenHash := sha256.Sum256([]byte(token))
				pm.EXPECT().CheckToken(ctx, nilUUID, tokenHash[:]).Return(false, errs.ErrIncorrectTokenBase)
			},
			expectedResult: errs.ErrIncorrectTokenBase,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			userRepo := mock_repositories.NewMockUserRepo(c)
			passwRepo := mock_repositories.NewMockPasswordResetRepo(c)
			templRepo := mock_repositories.NewMockTemplateRepo(c)
			transactor := mock_storage.NewMockTransactor(c)
			clouds := mock_cloudstorage.NewMockCloudStorage(c)
			log := logger.NewLogger(config.AppConfig{Name: "test", Env: "test", Version: "test", LogPath: ""})
			userServ := NewUserServ(userRepo, passwRepo, templRepo, clouds, transactor, log)
			ctx := context.Background()
			tt.mockBehavior(userRepo, transactor, passwRepo, ctx, tt.email, tt.token, tt.newPassword)
			res := userServ.ResetPassword(ctx, tt.email, tt.token, tt.newPassword)
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
