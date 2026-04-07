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

func (s *TaskService) Create(task *task) error {
	if task.Title == "" {
		return fmt.Errorf("title is required")
	}

	if task.Status == "" {
		return fmt.Errorf("invalid status")
	}

	task.ID = uuid.NewString()
	return s.repo.Create(task)
}

func (s *TaskService) GetAll() ([]task, error) {
	return s.repo.GetAll()
}
