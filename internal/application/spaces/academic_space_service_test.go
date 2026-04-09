package spaces_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

// STUBS

type stubPeriodRepo struct {
	period  *domain.AcademicPeriod
	periods []domain.AcademicPeriod
	err     error
}

func (s *stubPeriodRepo) FindByID(_ context.Context, _ int64) (*domain.AcademicPeriod, error) {
	return s.period, s.err
}
func (s *stubPeriodRepo) Create(_ context.Context, p *domain.AcademicPeriod) error {
	p.ID = 1
	return s.err
}
func (s *stubPeriodRepo) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	return s.periods, s.err
}
func (s *stubPeriodRepo) UpdateStatus(_ context.Context, _ int64, _ string) error {
	return s.err
}

type stubSpaceRepo struct {
	space  *domain.AcademicSpace
	spaces []domain.AcademicSpace
	err    error
}

func (s *stubSpaceRepo) Create(_ context.Context, sp *domain.AcademicSpace) error {
	sp.ID = 1
	sp.CreatedAt = time.Now()
	sp.UpdatedAt = time.Now()
	return s.err
}
func (s *stubSpaceRepo) FindByID(_ context.Context, _ int64) (*domain.AcademicSpace, error) {
	return s.space, s.err
}
func (s *stubSpaceRepo) FindByProfessor(_ context.Context, _ int64) ([]domain.AcademicSpace, error) {
	return s.spaces, s.err
}
func (s *stubSpaceRepo) UpdateStatus(_ context.Context, _ int64, _ string) error {
	return s.err
}

// HELPERS
func activePeriod() *domain.AcademicPeriod {
	return &domain.AcademicPeriod{ID: 1, Code: "2026-10", Status: "active"}
}

func closedPeriod() *domain.AcademicPeriod {
	return &domain.AcademicPeriod{ID: 2, Code: "2026-10", Status: "closed"}
}

func validSpaceInput(periodID, profID int64) spaces.CreateSpaceInput {
	return spaces.CreateSpaceInput{
		Name:             "Ingeniería de Software",
		Type:             domain.SpaceTypeCourse,
		AcademicPeriodID: periodID,
		ProfessorID:      profID,
		StartDate:        time.Now(),
		EndDate:          time.Now().Add(90 * 24 * time.Hour),
	}
}

func newSpaceSvc(spaceRepo domain.AcademicSpaceRepository, periodRepo domain.AcademicPeriodRepository) *spaces.AcademicSpaceService {
	return spaces.NewAcademicSpaceService(spaceRepo, periodRepo)
}



func TestCreateSpaceOK(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: activePeriod()})
	space, err := svc.CreateSpace(context.Background(), validSpaceInput(1, 10))
	if err != nil {
		t.Fatalf("esperaba nil error, obtuve: %v", err)
	}
	if space.ID == 0 {
		t.Error("esperaba ID asignado")
	}
}

//  no debe crear espacios en períodos cerrados 
func TestCreateSpacePeriodoCerrado(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: closedPeriod()})
	_, err := svc.CreateSpace(context.Background(), validSpaceInput(2, 10))
	if !errors.Is(err, domain.ErrPeriodoCerrado) {
		t.Errorf("esperaba ErrPeriodoCerrado, obtuve: %v", err)
	}
}

// período no encontrado.
func TestCreateSpacePeriodoNoEncontrado(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{err: domain.ErrPeriodoNoEncontrado})
	_, err := svc.CreateSpace(context.Background(), validSpaceInput(99, 10))
	if !errors.Is(err, domain.ErrPeriodoNoEncontrado) {
		t.Errorf("esperaba ErrPeriodoNoEncontrado, obtuve: %v", err)
	}
}

// tipo inválido.
func TestCreateSpaceTipoInvalido(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: activePeriod()})
	in := validSpaceInput(1, 10)
	in.Type = "seminar"
	_, err := svc.CreateSpace(context.Background(), in)
	if !errors.Is(err, domain.ErrTipoEspacioInvalido) {
		t.Errorf("esperaba ErrTipoEspacioInvalido, obtuve: %v", err)
	}
}

// Fechas incoherentes.
func TestCreateSpaceFechasInvalidas(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: activePeriod()})
	in := validSpaceInput(1, 10)
	in.EndDate = in.StartDate.Add(-time.Hour)
	_, err := svc.CreateSpace(context.Background(), in)
	if !errors.Is(err, domain.ErrFechasCierreInvalidas) {
		t.Errorf("esperaba ErrFechasCierreInvalidas, obtuve: %v", err)
	}
}

//  un profesor no puede ver el espacio de otro 
func TestGetSpaceOtroProfesor(t *testing.T) {
	svc := newSpaceSvc(
		&stubSpaceRepo{space: &domain.AcademicSpace{ID: 1, ProfessorID: 10, Status: "active"}},
		&stubPeriodRepo{},
	)
	_, err := svc.GetSpace(context.Background(), 1, 99)
	if !errors.Is(err, domain.ErrProfesorNoAutorizado) {
		t.Errorf("esperaba ErrProfesorNoAutorizado, obtuve: %v", err)
	}
}

//  cerrar un espacio ya cerrado debe fallar 
func TestCloseSpaceYaCerrado(t *testing.T) {
	svc := newSpaceSvc(
		&stubSpaceRepo{space: &domain.AcademicSpace{ID: 1, ProfessorID: 10, Status: "closed"}},
		&stubPeriodRepo{},
	)
	err := svc.CloseSpace(context.Background(), 1, 10)
	if !errors.Is(err, domain.ErrEspacioCerrado) {
		t.Errorf("esperaba ErrEspacioCerrado, obtuve: %v", err)
	}
}