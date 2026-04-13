package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/middleware"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type userRepoWithCreds struct {
	creds *domain.UserCredentials
}

func (u *userRepoWithCreds) FindCredentialsByEmail(_ context.Context, email string) (*domain.UserCredentials, error) {
	if u.creds == nil {
		return nil, nil
	}
	if email != u.creds.Email {
		return nil, nil
	}
	return u.creds, nil
}
func (*userRepoWithCreds) CreateUser(_ context.Context, _, _, _ string, _ []string) (int64, error) {
	return 0, nil
}
func (*userRepoWithCreds) ListUsers(_ context.Context) ([]domain.User, error) { return nil, nil }
func (*userRepoWithCreds) EmailExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (*userRepoWithCreds) CountUsers(_ context.Context) (int64, error) { return 0, nil }

var _ ports.UserRepository = (*userRepoWithCreds)(nil)

func jwtFromLogin(t *testing.T, secret []byte, creds *domain.UserCredentials, password string) string {
	t.Helper()
	svc := &auth.LoginService{Users: &userRepoWithCreds{creds: creds}, Secret: secret}
	res, err := svc.Login(context.Background(), creds.Email, password)
	if err != nil {
		t.Fatal(err)
	}
	return res.Token
}

func TestAutenticar_EmptySecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", middleware.Autenticar(nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code %d, want 500", rec.Code)
	}
}

func TestAutenticar_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-jwt-secret-at-least-32-characters-long")
	r := gin.New()
	r.GET("/x", middleware.Autenticar(secret), func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code %d, want 401", rec.Code)
	}
}

func TestAutenticar_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-jwt-secret-at-least-32-characters-long")
	r := gin.New()
	r.GET("/x", middleware.Autenticar(secret), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code %d, want 401", rec.Code)
	}
}

func TestAutenticar_BearerHeaderWithoutSpace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-jwt-secret-at-least-32-characters-long")
	r := gin.New()
	r.GET("/x", middleware.Autenticar(secret), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearerno-space")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code %d, want 401", rec.Code)
	}
}

func TestAutenticar_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-jwt-secret-at-least-32-characters-long")
	hash, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	creds := &domain.UserCredentials{
		ID: 42, Name: "N", Email: "n@test.co", PasswordHash: string(hash), Roles: []string{domain.RolProfesor},
	}
	token := jwtFromLogin(t, secret, creds, "pw")

	r := gin.New()
	r.GET("/x", middleware.Autenticar(secret), func(c *gin.Context) {
		uid, _ := c.Get("authUserID")
		if uid.(int64) != 42 {
			t.Fatalf("uid %v", uid)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d, want 200", rec.Code)
	}
}

func TestExigeRol_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-jwt-secret-at-least-32-characters-long")
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	creds := &domain.UserCredentials{
		ID: 1, Email: "a@test.co", PasswordHash: string(hash), Roles: []string{domain.RolAdministrador},
	}
	token := jwtFromLogin(t, secret, creds, "pw")

	r := gin.New()
	r.GET("/x", middleware.Autenticar(secret), middleware.ExigeRol(domain.RolAdministrador), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d, want 200", rec.Code)
	}
}

func TestExigeRol_ForbiddenWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-jwt-secret-at-least-32-characters-long")
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	creds := &domain.UserCredentials{
		ID: 1, Email: "p@test.co", PasswordHash: string(hash), Roles: []string{domain.RolProfesor},
	}
	token := jwtFromLogin(t, secret, creds, "pw")

	r := gin.New()
	r.GET("/x", middleware.Autenticar(secret), middleware.ExigeRol(domain.RolAdministrador), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code %d, want 403", rec.Code)
	}
}

func TestExigeRol_NoRolesInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", middleware.ExigeRol(domain.RolAdministrador), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code %d, want 403", rec.Code)
	}
}

func TestExigeRol_RolesNotStringSlice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", func(c *gin.Context) {
		c.Set("authRoles", 123)
		c.Next()
	}, middleware.ExigeRol(domain.RolAdministrador), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code %d, want 403", rec.Code)
	}
}
