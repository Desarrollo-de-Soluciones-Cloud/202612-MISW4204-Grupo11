package ports

import (
	"context"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type UserRepository interface {
	FindCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error)
	CreateUser(ctx context.Context, name, email, passwordHash string, roleNames []string) (int64, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
}
