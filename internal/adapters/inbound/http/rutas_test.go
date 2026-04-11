package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"

	"github.com/gin-gonic/gin"
)

type fakePinger struct {
	err error
}

type fakeTaskRepo struct{}

func (f fakeTaskRepo) Create(task *domain.Task) error {
	return nil
}

func (f fakeTaskRepo) GetAll() ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (f fakeTaskRepo) GetByID(id string) (*domain.Task, error) {
	return nil, nil
}

func (f fakeTaskRepo) Update(task *domain.Task) error {
	return nil
}

func (f fakeTaskRepo) Delete(id string) error {
	return nil
}

func (f fakeTaskRepo) SaveAttachment(attachment *domain.Attachment) error {
	return nil
}

func (f fakeTaskRepo) UpdateStatus(task *domain.Task) error {
	return nil
}

func (f fakePinger) Ping(_ context.Context) error {
	return f.err
}

func TestNuevoMotor_HealthAndTaskRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	readiness := &application.Readiness{DB: fakePinger{err: nil}}
	handler := handlers.NewTaskHandler(apptasks.NewTaskService(fakeTaskRepo{}))

	deps := Deps{
		Readiness:   readiness,
		JWTSecret:   []byte("test-secret"),
		Auth:        &handlers.Auth{},
		Users:       &handlers.Users{},
		TaskHandler: handler,
	}
	engine := NewEngine(deps)

	tests := []struct {
		name     string
		method   string
		path     string
		code     int
		expected string
	}{
		{name: "health", method: http.MethodGet, path: "/health", code: http.StatusOK, expected: "ok"},
		{name: "health ready", method: http.MethodGet, path: "/health/ready", code: http.StatusOK, expected: "ready"},
		{name: "tasks list", method: http.MethodGet, path: "/api/v1/tasks", code: http.StatusUnauthorized, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			engine.ServeHTTP(rr, req)

			if rr.Code != tt.code {
				t.Fatalf("expected status %d, got %d", tt.code, rr.Code)
			}

			if tt.expected == "ok" || tt.expected == "ready" {
				var payload map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if payload["status"] != tt.expected {
					t.Fatalf("expected status %q, got %q", tt.expected, payload["status"])
				}
			} else if tt.expected == "[]" {
				if rr.Body.String() != tt.expected {
					t.Fatalf("expected empty list %q, got %q", tt.expected, rr.Body.String())
				}
			}
		})
	}
}

func TestNuevoMotor_HealthReadyUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	readiness := &application.Readiness{DB: fakePinger{err: errTestPing}}
	handler := handlers.NewTaskHandler(apptasks.NewTaskService(nil))
	deps := Deps{
		Readiness:   readiness,
		JWTSecret:   []byte("test-secret"),
		Auth:        &handlers.Auth{},
		Users:       &handlers.Users{},
		TaskHandler: handler,
	}
	engine := NewEngine(deps)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
}

var errTestPing = &testPingError{}

type testPingError struct{}

func (e *testPingError) Error() string { return "ping failed" }
