package domain_test

import (
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

func TestAcademicSpace_IsActive(t *testing.T) {
	active := domain.AcademicSpace{Status: domain.SpaceStatusActive}
	if !active.IsActive() {
		t.Fatal("active space should be active")
	}
	closed := domain.AcademicSpace{Status: domain.SpaceStatusClosed}
	if closed.IsActive() {
		t.Fatal("closed space should not be active")
	}
	other := domain.AcademicSpace{Status: "other"}
	if other.IsActive() {
		t.Fatal("unknown status should not be active")
	}
}

func TestAcademicSpace_IsOpen(t *testing.T) {
	open := domain.AcademicSpace{Status: domain.SpaceStatusActive}
	if !open.IsOpen() {
		t.Fatal("IsOpen should match IsActive for active")
	}
	closed := domain.AcademicSpace{Status: domain.SpaceStatusClosed}
	if closed.IsOpen() {
		t.Fatal("closed space should not be open")
	}
}

func TestAcademicSpace_Validate_OK_Course(t *testing.T) {
	s := domain.AcademicSpace{
		Type:      domain.SpaceTypeCourse,
		StartDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAcademicSpace_Validate_OK_Project(t *testing.T) {
	s := domain.AcademicSpace{
		Type:      domain.SpaceTypeProject,
		StartDate: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAcademicSpace_Validate_TipoInvalido(t *testing.T) {
	s := domain.AcademicSpace{
		Type:      "seminar",
		StartDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
	}
	err := s.Validate()
	if err != domain.ErrTipoEspacioInvalido {
		t.Fatalf("want ErrTipoEspacioInvalido, got %v", err)
	}
}

func TestAcademicSpace_Validate_FechasInvalidas_EndBeforeStart(t *testing.T) {
	s := domain.AcademicSpace{
		Type:      domain.SpaceTypeCourse,
		StartDate: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	err := s.Validate()
	if err != domain.ErrFechasCierreInvalidas {
		t.Fatalf("want ErrFechasCierreInvalidas, got %v", err)
	}
}

func TestAcademicSpace_Validate_FechasInvalidas_SameDay(t *testing.T) {
	day := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	s := domain.AcademicSpace{
		Type:      domain.SpaceTypeCourse,
		StartDate: day,
		EndDate:   day,
	}
	err := s.Validate()
	if err != domain.ErrFechasCierreInvalidas {
		t.Fatalf("EndDate not After StartDate should fail, got %v", err)
	}
}
