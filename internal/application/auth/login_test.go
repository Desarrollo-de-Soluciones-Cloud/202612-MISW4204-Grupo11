package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type stubUserRepository struct {
	credentials *domain.UserCredentials
	findError   error
}

func (stub *stubUserRepository) FindCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error) {
	return stub.credentials, stub.findError
}

func (stub *stubUserRepository) CreateUser(ctx context.Context, name, email, passwordHash string, roleNames []string) (int64, error) {
	return 0, nil
}

func (stub *stubUserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	return nil, nil
}

func (stub *stubUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (stub *stubUserRepository) CountUsers(ctx context.Context) (int64, error) {
	return 0, nil
}

var _ ports.UserRepository = (*stubUserRepository)(nil)

func TestLogin_success(t *testing.T) {
	hash, hashErr := bcrypt.GenerateFromPassword([]byte("my-password-123"), bcrypt.DefaultCost)
	if hashErr != nil {
		t.Fatal(hashErr)
	}
	stub := &stubUserRepository{
		credentials: &domain.UserCredentials{
			ID:           1,
			Name:         "Ana",
			Email:        "ana@uniandes.edu.co",
			PasswordHash: string(hash),
			Roles:        []string{"profesor"},
		},
	}
	loginService := &auth.LoginService{
		Users:  stub,
		Secret: []byte("test-jwt-secret-at-least-32-characters"),
	}
	loginResult, loginErr := loginService.Login(context.Background(), "ana@uniandes.edu.co", "my-password-123")
	if loginErr != nil {
		t.Fatal(loginErr)
	}
	if loginResult.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if loginResult.User.ID != 1 || loginResult.User.Email != "ana@uniandes.edu.co" {
		t.Fatalf("unexpected user: %+v", loginResult.User)
	}
}

func TestLogin_invalidCredentials(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("other-secret"), bcrypt.DefaultCost)
	stub := &stubUserRepository{
		credentials: &domain.UserCredentials{
			ID:           1,
			Email:        "x@y.co",
			PasswordHash: string(hash),
		},
	}
	loginService := &auth.LoginService{
		Users:  stub,
		Secret: []byte("test-jwt-secret-at-least-32-characters"),
	}
	_, loginErr := loginService.Login(context.Background(), "x@y.co", "wrong-password")
	if loginErr == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(loginErr, auth.ErrInvalidCredentials) {
		t.Fatalf("unexpected error: %v", loginErr)
	}
}
