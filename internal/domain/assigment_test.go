package domain_test

import (
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

func TestAssignment_Validate_OK_Monitor(t *testing.T) {
	a := domain.Assignment{
		UserID:                 1,
		AcademicSpaceID:        2,
		ProfessorID:            10,
		RoleInAssignment:       domain.RoleMonitor,
		ContractedHoursPerWeek: 8,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAssignment_Validate_OK_GraduateAssistant(t *testing.T) {
	a := domain.Assignment{
		RoleInAssignment:       domain.RoleGraduateAssistant,
		ContractedHoursPerWeek: 10,
	}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAssignment_Validate_RolInvalido(t *testing.T) {
	a := domain.Assignment{
		RoleInAssignment:       "tutor",
		ContractedHoursPerWeek: 5,
	}
	err := a.Validate()
	if err != domain.ErrRolInvalido {
		t.Fatalf("want ErrRolInvalido, got %v", err)
	}
}

func TestAssignment_Validate_HorasCero(t *testing.T) {
	a := domain.Assignment{
		RoleInAssignment:       domain.RoleMonitor,
		ContractedHoursPerWeek: 0,
	}
	err := a.Validate()
	if err != domain.ErrHorasContratadas {
		t.Fatalf("want ErrHorasContratadas, got %v", err)
	}
}

func TestAssignment_Validate_HorasNegativas(t *testing.T) {
	a := domain.Assignment{
		RoleInAssignment:       domain.RoleGraduateAssistant,
		ContractedHoursPerWeek: -1,
	}
	err := a.Validate()
	if err != domain.ErrHorasContratadas {
		t.Fatalf("want ErrHorasContratadas, got %v", err)
	}
}

func TestAssignmentWithUser_EmbedsAssignment(t *testing.T) {
	aw := domain.AssignmentWithUser{
		Assignment: domain.Assignment{
			RoleInAssignment:       domain.RoleMonitor,
			ContractedHoursPerWeek: 4,
		},
		UserName:  "Pat",
		UserEmail: "pat@test.co",
	}
	if err := aw.Validate(); err != nil {
		t.Fatal(err)
	}
	if aw.UserName != "Pat" {
		t.Fatal("embedded fields should be accessible")
	}
}
