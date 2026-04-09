package domain

import (
	"errors"
	"time"
)

var ErrPeriodoNoEncontrado = errors.New("período académico no encontrado")

var ErrPeriodoCerrado = errors.New("el período académico está cerrado y no admite nuevos espacios")

type AcademicPeriod struct {
	ID        int64
	Code      string // 
	StartDate time.Time
	EndDate   time.Time
	Status    string // (active / closed)
	CreatedAt time.Time
}

func (p AcademicPeriod) IsOpen() bool {
	return p.Status == "active"
}
