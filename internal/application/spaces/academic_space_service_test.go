package spaces_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

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

func closedPeriod() *domain.AcademicPeriod {
	return &domain.AcademicPeriod{
		ID:        2,
		Code:      "2026-10",
		Status:    "closed",
		StartDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
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

func TestCreateSpacePeriodoCerrado(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: closedPeriod()})
	_, err := svc.CreateSpace(context.Background(), validSpaceInput(2, 10))
	if !errors.Is(err, domain.ErrPeriodoCerrado) {
		t.Errorf("esperaba ErrPeriodoCerrado, obtuve: %v", err)
	}
}

func TestCreateSpacePeriodoNoEncontrado(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{err: domain.ErrPeriodoNoEncontrado})
	_, err := svc.CreateSpace(context.Background(), validSpaceInput(99, 10))
	if !errors.Is(err, domain.ErrPeriodoNoEncontrado) {
		t.Errorf("esperaba ErrPeriodoNoEncontrado, obtuve: %v", err)
	}
}

func TestCreateSpaceTipoInvalido(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: activePeriod()})
	input := validSpaceInput(1, 10)
	input.Type = "seminar"
	_, err := svc.CreateSpace(context.Background(), input)
	if !errors.Is(err, domain.ErrTipoEspacioInvalido) {
		t.Errorf("esperaba ErrTipoEspacioInvalido, obtuve: %v", err)
	}
}

func TestCreateSpaceFechasInvalidas(t *testing.T) {
	svc := newSpaceSvc(&stubSpaceRepo{}, &stubPeriodRepo{period: activePeriod()})
	input := validSpaceInput(1, 10)
	input.EndDate = input.StartDate.Add(-time.Hour)
	_, err := svc.CreateSpace(context.Background(), input)
	if !errors.Is(err, domain.ErrFechasCierreInvalidas) {
		t.Errorf("esperaba ErrFechasCierreInvalidas, obtuve: %v", err)
	}
}

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

func TestListSpaces_DelegatesToRepository(t *testing.T) {
	want := []domain.AcademicSpace{{ID: 1, Name: "Aula", ProfessorID: 10}}
	svc := newSpaceSvc(
		&stubSpaceRepo{spaces: want},
		&stubPeriodRepo{},
	)
	got, err := svc.ListSpaces(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestListAllSpaces_DelegatesToRepository(t *testing.T) {
	want := []domain.AcademicSpace{{ID: 2, Name: "Lab"}}
	svc := newSpaceSvc(
		&stubSpaceRepo{spaces: want},
		&stubPeriodRepo{},
	)
	got, err := svc.ListAllSpaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Lab" {
		t.Fatalf("unexpected %+v", got)
	}
}
