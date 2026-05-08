package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type stubUserRepository struct {
	countUsersResult int64
	countUsersError  error

	emailExistsResult bool
	emailExistsError  error

	createUserID    int64
	createUserError error
	lastCreateName  string
	lastCreateEmail string
	lastCreateHash  string
	lastCreateRoles []string

	listUsersResult []domain.User
	listUsersError  error
}

func (stub *stubUserRepository) FindCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error) {
	return nil, nil
}

func (stub *stubUserRepository) CreateUser(ctx context.Context, name, email, passwordHash string, roleNames []string) (int64, error) {
	stub.lastCreateName = name
	stub.lastCreateEmail = email
	stub.lastCreateHash = passwordHash
	stub.lastCreateRoles = append([]string(nil), roleNames...)
	return stub.createUserID, stub.createUserError
}

func (stub *stubUserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	return stub.listUsersResult, stub.listUsersError
}

func (stub *stubUserRepository) ListUsersByRole(ctx context.Context, role string) ([]domain.User, error) {
	return stub.listUsersResult, stub.listUsersError
}

func (stub *stubUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return stub.emailExistsResult, stub.emailExistsError
}

func (stub *stubUserRepository) CountUsers(ctx context.Context) (int64, error) {
	return stub.countUsersResult, stub.countUsersError
}

var _ ports.UserRepository = (*stubUserRepository)(nil)

func TestAdminService_Create_success(t *testing.T) {
	stub := &stubUserRepository{
		emailExistsResult: false,
		createUserID:      99,
	}
	adminService := &users.AdminService{Users: stub}
	password := "valid-pass-8"
	created, err := adminService.Create(context.Background(), "  Ana  ", "  Ana@Example.ORG ", password, []string{"  PROFESOR ", "profesor"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 99 || created.Name != "Ana" || created.Email != "ana@example.org" {
		t.Fatalf("unexpected user: %+v", created)
	}
	if len(created.Roles) != 1 || created.Roles[0] != domain.RolProfesor {
		t.Fatalf("unexpected roles: %v", created.Roles)
	}
	if stub.lastCreateName != "Ana" || stub.lastCreateEmail != "ana@example.org" {
		t.Fatalf("repository got name=%q email=%q", stub.lastCreateName, stub.lastCreateEmail)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stub.lastCreateHash), []byte(password)); err != nil {
		t.Fatalf("stored hash does not match password: %v", err)
	}
	if len(stub.lastCreateRoles) != 1 || stub.lastCreateRoles[0] != domain.RolProfesor {
		t.Fatalf("repository roles: %v", stub.lastCreateRoles)
	}
}

func TestAdminService_Create_passwordTooShort(t *testing.T) {
	stub := &stubUserRepository{}
	adminService := &users.AdminService{Users: stub}
	_, err := adminService.Create(context.Background(), "x", "a@b.co", "short", []string{domain.RolProfesor})
	if !errors.Is(err, users.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestAdminService_Create_noRoles(t *testing.T) {
	stub := &stubUserRepository{}
	adminService := &users.AdminService{Users: stub}
	_, err := adminService.Create(context.Background(), "x", "a@b.co", "longenough", []string{"  ", ""})
	if !errors.Is(err, users.ErrNoRoles) {
		t.Fatalf("expected ErrNoRoles, got %v", err)
	}
}

func TestAdminService_Create_unknownRole(t *testing.T) {
	stub := &stubUserRepository{}
	adminService := &users.AdminService{Users: stub}
	_, err := adminService.Create(context.Background(), "x", "a@b.co", "longenough", []string{"ceo"})
	if !errors.Is(err, users.ErrUnknownRole) {
		t.Fatalf("expected ErrUnknownRole, got %v", err)
	}
}

func TestAdminService_Create_emailAlreadyRegistered(t *testing.T) {
	stub := &stubUserRepository{emailExistsResult: true}
	adminService := &users.AdminService{Users: stub}
	_, err := adminService.Create(context.Background(), "x", "taken@b.co", "longenough", []string{domain.RolProfesor})
	if !errors.Is(err, users.ErrEmailAlreadyRegistered) {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

func TestAdminService_Create_emailExistsError(t *testing.T) {
	repoErr := errors.New("db down")
	stub := &stubUserRepository{emailExistsError: repoErr}
	adminService := &users.AdminService{Users: stub}
	_, err := adminService.Create(context.Background(), "x", "a@b.co", "longenough", []string{domain.RolProfesor})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestAdminService_Create_createUserError(t *testing.T) {
	createErr := errors.New("insert failed")
	stub := &stubUserRepository{
		emailExistsResult: false,
		createUserError:   createErr,
	}
	adminService := &users.AdminService{Users: stub}
	_, err := adminService.Create(context.Background(), "x", "a@b.co", "longenough", []string{domain.RolMonitor})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestAdminService_List(t *testing.T) {
	expected := []domain.User{{ID: 1, Email: "a@b.co"}}
	stub := &stubUserRepository{listUsersResult: expected}
	adminService := &users.AdminService{Users: stub}
	list, err := adminService.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestAdminService_CountUsers(t *testing.T) {
	stub := &stubUserRepository{countUsersResult: 3}
	adminService := &users.AdminService{Users: stub}
	n, err := adminService.CountUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("expected 3, got %d", n)
	}
}
