package domain

import (
	"errors"
	"time"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskForbidden      = errors.New("forbidden: task does not belong to the current user")
	ErrAssignmentNotOwned = errors.New("assignment does not belong to the current user")
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
	Week           int       `json:"week"`
	TimeInvested   int       `json:"time_invested"`
	AssignmentId   int       `json:"assignment_id"`
	TimeRegistered time.Time `json:"time_registered"`
	Observations   string    `json:"observations"`
}
