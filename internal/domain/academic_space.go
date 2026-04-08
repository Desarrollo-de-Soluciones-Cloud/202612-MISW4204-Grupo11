package domain

import (
	"errors"
	"time"
)

// Tipos de espacio académico permitidos 
const (
	SpaceTypeCourse  = "course"
	SpaceTypeProject = "project"
)

// Estados de un espacio académico 
const (
	SpaceStatusActive = "active"
	SpaceStatusClosed = "closed"
)

var (
	ErrEspacioNoEncontrado     = errors.New("espacio académico no encontrado")
	ErrEspacioCerrado          = errors.New("el espacio académico está cerrado")
	ErrTipoEspacioInvalido     = errors.New("tipo de espacio inválido: debe ser 'course' o 'project'")
	ErrFechasCierreInvalidas   = errors.New("la fecha de cierre debe ser posterior a la de inicio")
	ErrProfesorNoAutorizado    = errors.New("no tienes permiso para operar sobre este espacio")
)

// AcademicSpace representa un curso o proyecto académico 
type AcademicSpace struct {
	ID               int64
	Name             string
	Type             string    // "course" "project"
	AcademicPeriodID int64
	ProfessorID      int64
	StartDate        time.Time
	EndDate          time.Time
	Observations     string
	Status           string    // "active" "closed"
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// informa si el espacio permite operaciones 
func (s AcademicSpace) IsActive() bool {
	return s.Status == SpaceStatusActive
}

//  verifica la coherencia básica de la entidad.
func (s AcademicSpace) Validate() error {
	if s.Type != SpaceTypeCourse && s.Type != SpaceTypeProject {
		return ErrTipoEspacioInvalido
	}
	if !s.EndDate.After(s.StartDate) {
		return ErrFechasCierreInvalidas
	}
	return nil
}