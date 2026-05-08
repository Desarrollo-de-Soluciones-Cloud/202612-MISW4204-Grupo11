package admin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type stubUserRepo struct {
	users []domain.User
	err   error
}

func (s *stubUserRepo) FindCredentialsByEmail(_ context.Context, _ string) (*domain.UserCredentials, error) {
	return nil, nil
}
func (s *stubUserRepo) CreateUser(_ context.Context, _, _, _ string, _ []string) (int64, error) {
	return 0, nil
}
func (s *stubUserRepo) ListUsers(_ context.Context) ([]domain.User, error) {
	return s.users, s.err
}
func (s *stubUserRepo) ListUsersByRole(_ context.Context, _ string) ([]domain.User, error) {
	return s.users, s.err
}
func (s *stubUserRepo) EmailExists(_ context.Context, _ string) (bool, error) { return false, nil }
func (s *stubUserRepo) CountUsers(_ context.Context) (int64, error)           { return 0, nil }

var _ ports.UserRepository = (*stubUserRepo)(nil)

type stubPeriodRepo struct {
	periods []domain.AcademicPeriod
	err     error
}

func (s *stubPeriodRepo) FindByID(_ context.Context, _ int64) (*domain.AcademicPeriod, error) {
	return nil, nil
}
func (s *stubPeriodRepo) Create(_ context.Context, _ *domain.AcademicPeriod) error { return nil }
func (s *stubPeriodRepo) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	return s.periods, s.err
}
func (s *stubPeriodRepo) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }

type stubSpaceRepo struct {
	spaces []domain.AcademicSpace
	err    error
}

func (s *stubSpaceRepo) Create(_ context.Context, _ *domain.AcademicSpace) error { return nil }
func (s *stubSpaceRepo) FindByID(_ context.Context, _ int64) (*domain.AcademicSpace, error) {
	return nil, nil
}
func (s *stubSpaceRepo) FindByProfessor(_ context.Context, _ int64) ([]domain.AcademicSpace, error) {
	return nil, nil
}
func (s *stubSpaceRepo) ListAll(_ context.Context) ([]domain.AcademicSpace, error) {
	return s.spaces, s.err
}
func (s *stubSpaceRepo) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }

type stubAssignRepo struct {
	assignments []domain.Assignment
	err         error
}

func (s *stubAssignRepo) Create(_ context.Context, _ *domain.Assignment) error { return nil }
func (s *stubAssignRepo) FindByID(_ context.Context, _ int64) (*domain.Assignment, error) {
	return nil, nil
}
func (s *stubAssignRepo) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (s *stubAssignRepo) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (s *stubAssignRepo) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return false, nil
}
func (s *stubAssignRepo) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}
func (s *stubAssignRepo) FindByProfessorWithUser(_ context.Context, _ int64) ([]domain.AssignmentWithUser, error) {
	return nil, nil
}
func (s *stubAssignRepo) ListAll(_ context.Context) ([]domain.Assignment, error) {
	return s.assignments, s.err
}
func (s *stubAssignRepo) Update(_ context.Context, _ *domain.Assignment) error { return nil }

type stubTaskLister struct {
	tasks []domain.Task
	err   error
}

func (s *stubTaskLister) ListAll(_ context.Context) ([]domain.Task, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tasks, s.err
}

func TestPlatformOverviewService_GetOverview_OK(t *testing.T) {
	users := []domain.User{{ID: 1, Name: "A", Email: "a@x.co", Roles: []string{"administrador"}}}
	periods := []domain.AcademicPeriod{{ID: 1, Code: "2026-1", Status: "active"}}
	spaces := []domain.AcademicSpace{{ID: 1, Name: "S", ProfessorID: 2}}
	assignments := []domain.Assignment{{ID: 1, UserID: 3, AcademicSpaceID: 1}}
	tasks := []domain.Task{{ID: 1, Title: "T", AssignmentId: 1}}

	svc := admin.NewPlatformOverviewService(
		&stubUserRepo{users: users},
		&stubPeriodRepo{periods: periods},
		&stubSpaceRepo{spaces: spaces},
		&stubAssignRepo{assignments: assignments},
		&stubTaskLister{tasks: tasks},
	)

	out, err := svc.GetOverview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Users) != 1 || len(out.AcademicPeriods) != 1 || len(out.AcademicSpaces) != 1 {
		t.Fatalf("unexpected counts: %+v", out)
	}
	if len(out.Assignments) != 1 || len(out.Tasks) != 1 {
		t.Fatalf("unexpected assignments/tasks")
	}
}

func TestPlatformOverviewService_GetOverview_PropagatesUserError(t *testing.T) {
	want := errors.New("db down")
	svc := admin.NewPlatformOverviewService(
		&stubUserRepo{err: want},
		&stubPeriodRepo{},
		&stubSpaceRepo{},
		&stubAssignRepo{},
		&stubTaskLister{},
	)
	_, err := svc.GetOverview(context.Background())
	if err == nil || !errors.Is(err, want) {
		t.Fatalf("expected wrapped want, got %v", err)
	}
}

func TestPlatformOverviewService_GetOverview_PropagatesPeriodsError(t *testing.T) {
	want := errors.New("periods fail")
	svc := admin.NewPlatformOverviewService(
		&stubUserRepo{users: []domain.User{{ID: 1}}},
		&stubPeriodRepo{err: want},
		&stubSpaceRepo{},
		&stubAssignRepo{},
		&stubTaskLister{},
	)
	_, err := svc.GetOverview(context.Background())
	if err == nil || !errors.Is(err, want) {
		t.Fatalf("expected periods error, got %v", err)
	}
}

func TestPlatformOverviewService_GetOverview_PropagatesTasksError(t *testing.T) {
	want := errors.New("tasks fail")
	svc := admin.NewPlatformOverviewService(
		&stubUserRepo{users: []domain.User{{ID: 1}}},
		&stubPeriodRepo{periods: []domain.AcademicPeriod{{ID: 1}}},
		&stubSpaceRepo{spaces: []domain.AcademicSpace{{ID: 1}}},
		&stubAssignRepo{assignments: []domain.Assignment{{ID: 1}}},
		&stubTaskLister{err: want},
	)
	_, err := svc.GetOverview(context.Background())
	if err == nil || !errors.Is(err, want) {
		t.Fatalf("expected tasks error, got %v", err)
	}
}
