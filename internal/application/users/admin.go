package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrNoRoles                = errors.New("at least one role is required")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrUnknownRole            = errors.New("unknown role")
)

type AdminService struct {
	Users ports.UserRepository
}

func (service *AdminService) CountUsers(ctx context.Context) (int64, error) {
	return service.Users.CountUsers(ctx)
}

func knownRoleNames() map[string]struct{} {
	return map[string]struct{}{
		domain.RolAdministrador:     {},
		domain.RolProfesor:          {},
		domain.RolMonitor:           {},
		domain.RolAsistenteGraduado: {},
	}
}

func normalizeRoleNames(names []string) []string {
	seen := make(map[string]struct{})
	var normalized []string
	for _, name := range names {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}

func (service *AdminService) Create(ctx context.Context, name, email, password string, roleNames []string) (domain.User, error) {
	var created domain.User
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	if len(password) < 8 {
		return created, ErrPasswordTooShort
	}
	roles := normalizeRoleNames(roleNames)
	if len(roles) == 0 {
		return created, ErrNoRoles
	}
	known := knownRoleNames()
	for _, roleName := range roles {
		if _, allowed := known[roleName]; !allowed {
			return created, fmt.Errorf("%w: %s", ErrUnknownRole, roleName)
		}
	}

	exists, existsErr := service.Users.EmailExists(ctx, email)
	if existsErr != nil {
		return created, existsErr
	}
	if exists {
		return created, ErrEmailAlreadyRegistered
	}

	passwordHash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return created, hashErr
	}

	userID, createErr := service.Users.CreateUser(ctx, name, email, string(passwordHash), roles)
	if createErr != nil {
		return created, createErr
	}

	created.ID = userID
	created.Name = name
	created.Email = email
	created.Roles = roles
	return created, nil
}

func (service *AdminService) List(ctx context.Context) ([]domain.User, error) {
	return service.Users.ListUsers(ctx)
}
