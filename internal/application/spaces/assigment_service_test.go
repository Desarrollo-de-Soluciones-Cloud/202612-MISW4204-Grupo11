package spaces_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type stubAssignmentRepo struct {
	assignment  *domain.Assignment
	assignments []domain.Assignment
	exists      bool
	err         error
}

func (stub *stubAssignmentRepo) Create(_ context.Context, assignment *domain.Assignment) error {
	assignment.ID = 1
	return stub.err
}
func (stub *stubAssignmentRepo) FindByID(_ context.Context, _ int64) (*domain.Assignment, error) {
	return stub.assignment, stub.err
}
func (stub *stubAssignmentRepo) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return stub.assignments, stub.err
}
func (stub *stubAssignmentRepo) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return stub.assignments, stub.err
}
func (stub *stubAssignmentRepo) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return stub.exists, stub.err
}
func (stub *stubAssignmentRepo) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return stub.assignments, stub.err
}

func (stub *stubAssignmentRepo) FindByProfessorWithUser(_ context.Context, _ int64) ([]domain.AssignmentWithUser, error) {
	return nil, stub.err
}

func (stub *stubAssignmentRepo) ListAll(_ context.Context) ([]domain.Assignment, error) {
	return stub.assignments, stub.err
}

type stubSpaceRepo struct {
	space  *domain.AcademicSpace
	spaces []domain.AcademicSpace
	err    error
}

func (stubSpace *stubSpaceRepo) Create(_ context.Context, space *domain.AcademicSpace) error {
	space.ID = 1
	return stubSpace.err
}
func (stubSpace *stubSpaceRepo) FindByID(_ context.Context, _ int64) (*domain.AcademicSpace, error) {
	return stubSpace.space, stubSpace.err
}
func (stubSpace *stubSpaceRepo) FindByProfessor(_ context.Context, _ int64) ([]domain.AcademicSpace, error) {
	return stubSpace.spaces, stubSpace.err
}
func (stubSpace *stubSpaceRepo) UpdateStatus(_ context.Context, _ int64, _ string) error {
	return stubSpace.err
}

func (stubSpace *stubSpaceRepo) ListAll(_ context.Context) ([]domain.AcademicSpace, error) {
	if len(stubSpace.spaces) > 0 {
		return stubSpace.spaces, stubSpace.err
	}
	if stubSpace.space != nil {
		return []domain.AcademicSpace{*stubSpace.space}, stubSpace.err
	}
	return nil, stubSpace.err
}

type stubPeriodRepo struct {
	period  *domain.AcademicPeriod
	periods []domain.AcademicPeriod
	err     error
}

func (stubPeriod *stubPeriodRepo) FindByID(_ context.Context, _ int64) (*domain.AcademicPeriod, error) {
	return stubPeriod.period, stubPeriod.err
}
func (stubPeriod *stubPeriodRepo) Create(_ context.Context, period *domain.AcademicPeriod) error {
	period.ID = 1
	return stubPeriod.err
}
func (stubPeriod *stubPeriodRepo) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	return stubPeriod.periods, stubPeriod.err
}
func (stubPeriod *stubPeriodRepo) UpdateStatus(_ context.Context, _ int64, _ string) error {
	return stubPeriod.err
}

func activePeriod() *domain.AcademicPeriod {
	return &domain.AcademicPeriod{
		ID:        1,
		Code:      "2024-1",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		Status:    "active",
	}
}

func activeSpace() *domain.AcademicSpace {
	return &domain.AcademicSpace{
		ID:               1,
		ProfessorID:      10,
		AcademicPeriodID: 1,
		Status:           "active",
		StartDate:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC),
	}
}

func closedSpace() *domain.AcademicSpace {
	return &domain.AcademicSpace{
		ID:               1,
		ProfessorID:      10,
		AcademicPeriodID: 1,
		Status:           "closed",
		StartDate:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC),
	}
}

func unauthorizedProfessorSpace() *domain.AcademicSpace {
	return &domain.AcademicSpace{
		ID:               1,
		ProfessorID:      99,
		AcademicPeriodID: 1,
		Status:           "active",
		StartDate:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC),
	}
}

func validAssignmentInput() appspaces.CreateAssignmentInput {
	return appspaces.CreateAssignmentInput{
		UserID:                 5,
		AcademicSpaceID:        1,
		ProfessorID:            10,
		RoleInAssignment:       domain.RoleMonitor,
		ContractedHoursPerWeek: 8,
	}
}

func newAssignmentSvc(spaceRepo domain.AcademicSpaceRepository, assignRepo domain.AssignmentRepository) *appspaces.AssignmentService {
	return appspaces.NewAssignmentService(assignRepo, spaceRepo, &stubPeriodRepo{period: activePeriod()}, appspaces.NoOpHourRuleChecker{})
}

func (stub *stubAssignmentRepo) Update(_ context.Context, assignment *domain.Assignment) error {
	return stub.err
}

func TestCreateAssignmentOK(t *testing.T) {
	svc := newAssignmentSvc(
		&stubSpaceRepo{space: activeSpace()},
		&stubAssignmentRepo{},
	)
	assignment, err := svc.CreateAssignment(context.Background(), validAssignmentInput())
	if err != nil {
		t.Fatalf("esperaba nil error, obtuve: %v", err)
	}
	if assignment.ID == 0 {
		t.Error("esperaba ID asignado")
	}
}

func TestCreateAssignmentEspacioCerrado(t *testing.T) {
	svc := newAssignmentSvc(&stubSpaceRepo{space: closedSpace()}, &stubAssignmentRepo{})
	_, err := svc.CreateAssignment(context.Background(), validAssignmentInput())
	if !errors.Is(err, domain.ErrEspacioCerradoVinculacion) {
		t.Errorf("esperaba ErrEspacioCerradoVinculacion, obtuve: %v", err)
	}
}

func TestCreateAssignmentProfesorNoAutorizado(t *testing.T) {
	svc := newAssignmentSvc(
		&stubSpaceRepo{space: unauthorizedProfessorSpace()},
		&stubAssignmentRepo{},
	)
	_, err := svc.CreateAssignment(context.Background(), validAssignmentInput())
	if !errors.Is(err, domain.ErrProfesorNoAutorizado) {
		t.Errorf("esperaba ErrProfesorNoAutorizado, obtuve: %v", err)
	}
}

func TestCreateAssignmentDuplicado(t *testing.T) {
	svc := newAssignmentSvc(
		&stubSpaceRepo{space: activeSpace()},
		&stubAssignmentRepo{exists: true},
	)
	_, err := svc.CreateAssignment(context.Background(), validAssignmentInput())
	if !errors.Is(err, domain.ErrUsuarioYaVinculado) {
		t.Errorf("esperaba ErrUsuarioYaVinculado, obtuve: %v", err)
	}
}

func TestCreateAssignmentRolInvalido(t *testing.T) {
	svc := newAssignmentSvc(&stubSpaceRepo{space: activeSpace()}, &stubAssignmentRepo{})
	input := validAssignmentInput()
	input.RoleInAssignment = "tutor"
	_, err := svc.CreateAssignment(context.Background(), input)
	if !errors.Is(err, domain.ErrRolInvalido) {
		t.Errorf("esperaba ErrRolInvalido, obtuve: %v", err)
	}
}

func TestCreateAssignmentHorasInvalidas(t *testing.T) {
	svc := newAssignmentSvc(&stubSpaceRepo{space: activeSpace()}, &stubAssignmentRepo{})
	input := validAssignmentInput()
	input.ContractedHoursPerWeek = 0
	_, err := svc.CreateAssignment(context.Background(), input)
	if !errors.Is(err, domain.ErrHorasContratadas) {
		t.Errorf("esperaba ErrHorasContratadas, obtuve: %v", err)
	}
}

func TestCreateAssignmentPeriodoCerrado(t *testing.T) {
	closedPeriod := &domain.AcademicPeriod{
		ID:        1,
		Code:      "2024-1",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		Status:    "closed",
	}

	svc := appspaces.NewAssignmentService(
		&stubAssignmentRepo{},
		&stubSpaceRepo{space: activeSpace()},
		&stubPeriodRepo{period: closedPeriod},
		appspaces.NoOpHourRuleChecker{},
	)
	_, err := svc.CreateAssignment(context.Background(), validAssignmentInput())
	if !errors.Is(err, domain.ErrPeriodoCerradoVinculacion) {
		t.Errorf("esperaba ErrPeriodoCerradoVinculacion, obtuve: %v", err)
	}
}

func TestCreateAssignmentFechasEspacioFueraDelPeriodo(t *testing.T) {
	tests := []struct {
		name  string
		space *domain.AcademicSpace
	}{
		{
			name: "Espacio inicia antes del periodo",
			space: &domain.AcademicSpace{
				ID:               1,
				ProfessorID:      10,
				AcademicPeriodID: 1,
				Status:           "active",
				StartDate:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:          time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Espacio termina después del periodo",
			space: &domain.AcademicSpace{
				ID:               1,
				ProfessorID:      10,
				AcademicPeriodID: 1,
				Status:           "active",
				StartDate:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				EndDate:          time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			svc := appspaces.NewAssignmentService(
				&stubAssignmentRepo{},
				&stubSpaceRepo{space: testCase.space},
				&stubPeriodRepo{period: activePeriod()},
				appspaces.NoOpHourRuleChecker{},
			)
			_, err := svc.CreateAssignment(context.Background(), validAssignmentInput())
			if !errors.Is(err, domain.ErrFechasEspacioFueraDelPeriodo) {
				t.Errorf("esperaba ErrFechasEspacioFueraDelPeriodo, obtuve: %v", err)
			}
		})
	}
}
