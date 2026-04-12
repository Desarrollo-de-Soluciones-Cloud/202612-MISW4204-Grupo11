package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/adapters/inbound/http/handlers"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application"
	appadmin "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"

	"github.com/gin-gonic/gin"
)

type fakePinger struct {
	err error
}

type fakeAssignmentRepoForRoutes struct{}

func (fakeAssignmentRepoForRoutes) Create(_ context.Context, _ *domain.Assignment) error { return nil }
func (fakeAssignmentRepoForRoutes) FindByID(_ context.Context, _ int64) (*domain.Assignment, error) {
	return nil, domain.ErrVinculacionNoEncontrada
}
func (fakeAssignmentRepoForRoutes) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (fakeAssignmentRepoForRoutes) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (fakeAssignmentRepoForRoutes) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return false, nil
}
func (fakeAssignmentRepoForRoutes) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}
func (fakeAssignmentRepoForRoutes) FindByProfessorWithUser(_ context.Context, _ int64) ([]domain.AssignmentWithUser, error) {
	return nil, nil
}
func (fakeAssignmentRepoForRoutes) Update(_ context.Context, _ *domain.Assignment) error { return nil }

func (fakeAssignmentRepoForRoutes) ListAll(_ context.Context) ([]domain.Assignment, error) {
	return nil, nil
}

type stubUserRepoForRoutes struct{}

func (stubUserRepoForRoutes) FindCredentialsByEmail(_ context.Context, _ string) (*domain.UserCredentials, error) {
	return nil, errors.New("stub")
}
func (stubUserRepoForRoutes) CreateUser(_ context.Context, _, _, _ string, _ []string) (int64, error) {
	return 0, nil
}
func (stubUserRepoForRoutes) ListUsers(_ context.Context) ([]domain.User, error) {
	return nil, nil
}
func (stubUserRepoForRoutes) EmailExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (stubUserRepoForRoutes) CountUsers(_ context.Context) (int64, error) {
	return 0, nil
}

var _ ports.UserRepository = stubUserRepoForRoutes{}

type fakePeriodRepoForRoutes struct{}

func (fakePeriodRepoForRoutes) FindByID(_ context.Context, _ int64) (*domain.AcademicPeriod, error) {
	return nil, nil
}
func (fakePeriodRepoForRoutes) Create(_ context.Context, _ *domain.AcademicPeriod) error { return nil }
func (fakePeriodRepoForRoutes) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	return nil, nil
}
func (fakePeriodRepoForRoutes) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }

type fakeSpaceRepoForRoutes struct{}

func (fakeSpaceRepoForRoutes) Create(_ context.Context, _ *domain.AcademicSpace) error { return nil }
func (fakeSpaceRepoForRoutes) FindByID(_ context.Context, _ int64) (*domain.AcademicSpace, error) {
	return nil, nil
}
func (fakeSpaceRepoForRoutes) FindByProfessor(_ context.Context, _ int64) ([]domain.AcademicSpace, error) {
	return nil, nil
}
func (fakeSpaceRepoForRoutes) ListAll(_ context.Context) ([]domain.AcademicSpace, error) {
	return nil, nil
}
func (fakeSpaceRepoForRoutes) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }

type fakeTaskRepo struct{}

func (f fakeTaskRepo) Create(task *domain.Task) error {
	return nil
}

func (f fakeTaskRepo) ListAll(ctx context.Context) ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (f fakeTaskRepo) ListByUser(ctx context.Context, userID int64) ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (f fakeTaskRepo) ListByProfessorID(ctx context.Context, professorID int64) ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (f fakeTaskRepo) GetByID(id string) (*domain.Task, error) {
	return nil, nil
}

func (f fakeTaskRepo) GetByIDForUser(ctx context.Context, id string, userID int64) (*domain.Task, error) {
	return nil, domain.ErrTaskNotFound
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

func testAdminHandler() *handlers.Admin {
	overview := appadmin.NewPlatformOverviewService(
		stubUserRepoForRoutes{},
		fakePeriodRepoForRoutes{},
		fakeSpaceRepoForRoutes{},
		fakeAssignmentRepoForRoutes{},
		fakeTaskRepo{},
	)
	return handlers.NewAdminHandler(overview)
}

func TestNuevoMotor_HealthAndTaskRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	readiness := &application.Readiness{DB: fakePinger{err: nil}}
	handler := handlers.NewTaskHandler(apptasks.NewTaskService(fakeTaskRepo{}, fakeAssignmentRepoForRoutes{}))

	deps := Deps{
		Readiness:   readiness,
		JWTSecret:   []byte("test-secret"),
		Auth:        &handlers.Auth{},
		Users:       &handlers.Users{},
		Admin:       testAdminHandler(),
		TaskHandler: handler,
		AcadSpaces:  &handlers.AcademicSpaceHandler{},
		Periods:     &handlers.AcademicPeriodHandler{},
		Assignments: &handlers.AssignmentHandler{},
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
	handler := handlers.NewTaskHandler(apptasks.NewTaskService(nil, fakeAssignmentRepoForRoutes{}))
	deps := Deps{
		Readiness:   readiness,
		JWTSecret:   []byte("test-secret"),
		Auth:        &handlers.Auth{},
		Users:       &handlers.Users{},
		Admin:       testAdminHandler(),
		TaskHandler: handler,
		AcadSpaces:  &handlers.AcademicSpaceHandler{},
		Periods:     &handlers.AcademicPeriodHandler{},
		Assignments: &handlers.AssignmentHandler{},
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
