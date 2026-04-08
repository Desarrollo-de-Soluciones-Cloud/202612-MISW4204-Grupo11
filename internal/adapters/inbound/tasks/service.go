package tasks

import (
	"fmt"

	"github.com/google/uuid"
)

type TaskService struct {
	repo Repository
}

func NewTaskService(repo Repository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(task *task, user *User) error {
	if task.Title == "" {
		return fmt.Errorf("title is required")
	}

	if task.Status == "" {
		return fmt.Errorf("invalid status")
	}

	if limitOfTimeRecorded22(user, task.Week, task.TimeInvested) == true {
		return fmt.Errorf("No se pueden registrar mas horas, ya se paso las 22 horas.")
	}

	task.ID = uuid.NewString()
	return s.repo.Create(task)
}

func (s *TaskService) GetAll() ([]task, error) {
	return s.repo.GetAll()
}

func limitOfTimeRecorded22(user *User, week int, newTaskHours int) bool {
	total := 0

	for _, vinculation := range user.Vinculations {
		if vinculation.Role != "assistant_graduated" {
			continue
		}

		for _, task := range vinculation.Tasks {
			if task.Week == week {
				total += task.TimeInvested
			}
		}
	}

	total += newTaskHours

	if total > 22 {
		return true
	}
	return false
}
