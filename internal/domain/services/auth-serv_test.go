package services

// import (
// 	"context"
// 	mock_repositories "readmeow/internal/domain/repositories/mocks"
// 	mock_cloudstorage "readmeow/pkg/cloudstorage/mocks"
// 	mock_storage "readmeow/pkg/storage/mocks"
// 	"testing"
// 	"time"

// 	"github.com/golang/mock/gomock"
// 	"github.com/google/uuid"
// )

// func TestAuthServ_Register(t *testing.T) {
// 	nilUUID := uuid.Nil
// 	type mockBehavior = func(um *mock_repositories.MockUserRepo, vm *mock_repositories.MockVerificationRepo, tm *mock_storage.MockTransactor, cm *mock_cloudstorage.MockCloudStorage, ctx context.Context, email, code string)
// 	tests := []struct {
// 		name           string
// 		email          string
// 		code           string
// 		mockBehavior   mockBehavior
// 		expectedResult *loginData
// 		expectedError  error
// 	}{
// 		{
// 			name:  "success",
// 			email: "email",
// 			code:  nilUUID.String(),
// 			mockBehavior: func(um *mock_repositories.MockUserRepo, vm *mock_repositories.MockVerificationRepo, tm *mock_storage.MockTransactor, cm *mock_cloudstorage.MockCloudStorage, ctx context.Context, email, code string) {
// 				tm.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(func(txCtx context.Context, f func(context.Context) (any, error)) (any, error) {
// 					_, err := f(txCtx)
// 					return nil, err
// 				})
// 			},
// 			expectedResult: &loginData{
// 				Id:       nilUUID.String(),
// 				Nickname: "nick",
// 				Avatar:   "avatar",
// 				JWT:      "jwt",
// 				TTL:      time.Now(),
// 			},
// 			expectedError: nil,
// 		},
// 	}
// }
