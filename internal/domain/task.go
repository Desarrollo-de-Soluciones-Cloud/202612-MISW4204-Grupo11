package domain

import "time"

type Status string

const (
	StatusOpen          Status = "Abierto"
	StatusInDevelopment Status = "En desarrollo"
	StatusFinalized     Status = "finalizado"
)

type Task struct {
	ID             int
	Title          string
	Description    string
	Status         Status
	Week           int
	TimeInvested   int
	TimeRegistered time.Time
	Observations   string
}
