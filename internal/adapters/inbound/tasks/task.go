package tasks

import "time"

type Status string

const (
	StatusOpen          Status = "Abierto"
	StatusInDevelopment Status = "En desarrollo"
	StatusFinalized     Status = "finalizado"
)

type Vinculation struct {
	Role  string
	Tasks []task
}

type User struct {
	Username     string
	Password     string
	Vinculations []Vinculation
}

type task struct {
	ID             string
	Title          string
	Description    string
	Status         Status
	Week           int
	TimeInvested   int
	TimeRegistered time.Time
	Observations   string
}

func newTask(title string, description string, status Status, week int, timeInvested int, observations string) *task {
	return &task{
		Title:        title,
		Description:  description,
		Status:       status,
		Week:         week,
		TimeInvested: timeInvested,
		Observations: observations,
	}
}
