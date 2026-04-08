package domain

import (
	"errors"
	"time"
)

//  devuelve cuando no existe el período académico.
var ErrPeriodoNoEncontrado = errors.New("período académico no encontrado")

//  devuelve cuando se intenta operar sobre un período cerrado.
var ErrPeriodoCerrado = errors.New("el período académico está cerrado y no admite nuevos espacios")

// AcademicPeriod representa un período académico 
type AcademicPeriod struct {
	ID        int64
	Code      string // e.g. "2026-10"
	StartDate time.Time
	EndDate   time.Time
	Status    string // "active" | "closed"
}

// informa si el período está abierto para creación de espacios.
func (p AcademicPeriod) IsOpen() bool {
	return p.Status == "active"
}
