package domain

import (
	"errors"
	"time"
)

var (
	ErrReporteNoEncontrado   = errors.New("reporte no encontrado")
	ErrReporteNoAutorizado   = errors.New("no tienes permiso para acceder a este reporte")
	ErrGeneracionIAFallida   = errors.New("no se pudo generar el resumen con IA")
	ErrGeneracionPDFFallida  = errors.New("no se pudo generar el archivo PDF")
)

type Report struct {
	ID           int64     `json:"id"`
	ProfessorID  int64     `json:"professor_id"`
	AssignmentID int64     `json:"assignment_id"`
	UserName     string    `json:"user_name"`
	UserEmail    string    `json:"user_email"`
	Role         string    `json:"role"`
	WeekStart    time.Time `json:"week_start"`
	FilePath     string    `json:"-"`
	AISummary    string    `json:"ai_summary,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
