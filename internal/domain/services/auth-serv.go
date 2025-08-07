package services

// import (
// 	"context"
// 	"log/slog"
// 	"readmeow/internal/domain/models"
// 	"readmeow/internal/domain/repositories"
// )

// type AuthServ interface {
// 	Register(ctx context.Context, login, email, password string) error
// 	Login(ctx context.Context, login, password string) (string, error)
// 	Profile(ctx context.Context) (*models.User, error)
// }

// type authServ struct {
// 	UserRepo repositories.UserRepo
// 	Logger   *slog.Logger
// }

// func NewAuthServ(ur repositories.UserRepo, l *slog.Logger) AuthServ {
// 	return &authServ{
// 		UserRepo: ur,
// 		Logger:   l,
// 	}
// }

// func (ur *authServ) Register(ctx context.Context, login, email, password string) error {

// }

// func (ur *authServ) Login(ctx context.Context, login, password string) error {

// }

// func (ur *authServ) Profile(ctx context.Context) (*models.User, error) {

// }
