package domain

import (
	"errors"
	"time"
)

const (
	SpaceTypeCourse  = "course"
	SpaceTypeProject = "project"
)

const (
	SpaceStatusActive = "active"
	SpaceStatusClosed = "closed"
)

var (
	ErrEspacioNoEncontrado   = errors.New("espacio académico no encontrado")
	ErrEspacioCerrado        = errors.New("el espacio académico está cerrado")
	ErrTipoEspacioInvalido   = errors.New("tipo de espacio inválido: debe ser 'course' o 'project'")
	ErrFechasCierreInvalidas = errors.New("la fecha de cierre debe ser posterior a la de inicio")
	ErrProfesorNoAutorizado  = errors.New("no tienes permiso para operar sobre este espacio")
)

type AcademicSpace struct {
	ID               int64
	Name             string
	Type             string
	AcademicPeriodID int64
	ProfessorID      int64
	StartDate        time.Time
	EndDate          time.Time
	Observations     string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (space AcademicSpace) IsActive() bool {
	return space.Status == SpaceStatusActive
}

func (space AcademicSpace) IsOpen() bool {
	return space.IsActive()
}

func (space AcademicSpace) Validate() error {
	if space.Type != SpaceTypeCourse && space.Type != SpaceTypeProject {
		return ErrTipoEspacioInvalido
	}
	if !space.EndDate.After(space.StartDate) {
		return ErrFechasCierreInvalidas
	}
	return nil
}
