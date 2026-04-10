package domain

import "time"

type Status string

const (
	StatusOpen          Status = "Abierto"
	StatusInDevelopment Status = "En desarrollo"
	StatusFinalized     Status = "finalizado"
)

type Vinculation struct {
	Role  string
	Tasks []Task
}

/*type User struct {
	Username     string
	Password     string
	Vinculations []Vinculation
}*/

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

func newTask(title string, description string, status Status, week int, timeInvested int, observations string) *Task {
	return &Task{
		Title:        title,
		Description:  description,
		Status:       status,
		Week:         week,
		TimeInvested: timeInvested,
		Observations: observations,
	}
}
