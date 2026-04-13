package domain

import (
	"errors"
	"time"
)

var (
	ErrTaskNotFound              = errors.New("task not found")
	ErrTaskForbidden             = errors.New("forbidden: task does not belong to the current user")
	ErrAssignmentNotOwned        = errors.New("assignment does not belong to the current user")
	ErrSemanaInicioNoEsLunes     = errors.New("week_start debe ser un lunes")
	ErrSemanaFutura              = errors.New("no se pueden registrar tareas para semanas futuras")
	ErrModificacionFueraDeSemana = errors.New("solo se pueden modificar tareas de la semana activa")
	ErrEliminacionFueraDeSemana  = errors.New("solo se pueden eliminar tareas de la semana activa")
	ErrReporteTardioInmutable    = errors.New("los reportes tardíos no pueden ser modificados")
	ErrReporteTardioNoEliminable = errors.New("los reportes tardíos no pueden ser eliminados")
)

type Status string

const (
	StatusOpen          Status = "Abierto"
	StatusInDevelopment Status = "En desarrollo"
	StatusFinalized     Status = "finalizado"
)

type Task struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Status         Status    `json:"status"`
	WeekStart      time.Time `json:"week_start"`
	IsLate         bool      `json:"is_late"`
	TimeInvested   int       `json:"time_invested"`
	AssignmentId   int       `json:"assignment_id"`
	TimeRegistered time.Time `json:"time_registered"`
	Observations   string    `json:"observations"`
}

// BelongsToCurrentWeek returns true if the task's WeekStart matches the current week's Monday.
func (t Task) BelongsToCurrentWeek() bool {
	return IsCurrentWeek(t.WeekStart)
}

// CanBeModified returns true if the task is in the current week and is not a late report.
func (t Task) CanBeModified() bool {
	return t.BelongsToCurrentWeek() && !t.IsLate
}
