package domain

import (
	"errors"
	"time"
)

const (
	RoleMonitor           = "monitor"
	RoleGraduateAssistant = "graduate_assistant"
)

var (
	ErrVinculacionNoEncontrada       = errors.New("vinculación no encontrada")
	ErrRolInvalido                   = errors.New("rol inválido: debe ser 'monitor' o 'graduate_assistant'")
	ErrHorasContratadas              = errors.New("las horas contratadas deben ser un entero positivo")
	ErrEspacioCerradoVinculacion     = errors.New("no se puede vincular a un espacio académico cerrado")
	ErrUsuarioYaVinculado            = errors.New("el usuario ya tiene una vinculación activa en este espacio con el mismo rol")
	ErrPeriodoCerradoVinculacion     = errors.New("no se puede vincular dentro de un período académico cerrado")
	ErrFechasEspacioFueraDelPeriodo  = errors.New("las fechas del espacio académico deben estar dentro del período académico")
	
)

type Assignment struct {
	ID                     int64
	UserID                 int64
	AcademicSpaceID        int64
	ProfessorID            int64
	RoleInAssignment       string
	ContractedHoursPerWeek int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type AssignmentWithUser struct {
	Assignment
	UserName  string
	UserEmail string
}

func (assignment Assignment) Validate() error {
	if assignment.RoleInAssignment != RoleMonitor && assignment.RoleInAssignment != RoleGraduateAssistant {
		return ErrRolInvalido
	}
	if assignment.ContractedHoursPerWeek <= 0 {
		return ErrHorasContratadas
	}
	return nil
}