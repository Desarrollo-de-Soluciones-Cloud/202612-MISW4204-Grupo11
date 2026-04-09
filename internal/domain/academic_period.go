package domain

import (
	"errors"
	"time"
)

var ErrPeriodoNoEncontrado = errors.New("período académico no encontrado")

var ErrPeriodoCerrado = errors.New("el período académico está cerrado y no admite nuevos espacios")

type AcademicPeriod struct {
	ID        int64
	Code      string 
	StartDate time.Time
	EndDate   time.Time
	Status    string 
	CreatedAt time.Time
}

func (period AcademicPeriod) IsOpen() bool {
	return period.Status == "active"
}
