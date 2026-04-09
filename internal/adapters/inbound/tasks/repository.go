package tasks

import (
	"fmt"
)

type Repository interface {
	Create(task *task) error
	GetAll() ([]task, error)
	GetByID(id string) (*task, error)
	Update(task *task) error
	Delete(id string) error
	SaveAttachment(attachment *Attachment) error
	UpdateStatus(task *task) error
}

type TaskRepository struct {
	tasks       []task
	nextID      int
	attachments []Attachment
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks:  []task{},
		nextID: 1,
	}
}

func (r *TaskRepository) Create(task *task) error {
	task.ID = r.nextID
	r.nextID++

	r.tasks = append(r.tasks, *task)
	return nil
}

func (r *TaskRepository) GetAll() ([]task, error) {
	return r.tasks, nil
}

func (r *TaskRepository) GetByID(id string) (*task, error) {
	// Parse id to int
	var intID int
	if _, err := fmt.Sscanf(id, "%d", &intID); err != nil {
		return nil, fmt.Errorf("invalid id")
	}
	for i := range r.tasks {
		if r.tasks[i].ID == intID {
			return &r.tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task not found")
}

func (r *TaskRepository) Update(task *task) error {
	for i := range r.tasks {
		if r.tasks[i].ID == task.ID {
			r.tasks[i] = *task
			return nil
		}
	}
	return fmt.Errorf("task no se ha encontrado")
}

func (r *TaskRepository) Delete(id string) error {
	var intID int
	if _, err := fmt.Sscanf(id, "%d", &intID); err != nil {
		return fmt.Errorf("invalid id")
	}
	for i := range r.tasks {
		if r.tasks[i].ID == intID {
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task not found")
}

func (r *TaskRepository) SaveAttachment(attachment *Attachment) error {
	r.attachments = append(r.attachments, *attachment)
	return nil
}

func (r *TaskRepository) UpdateStatus(task *task) error {
	for i := range r.tasks {
		if r.tasks[i].ID == task.ID {
			r.tasks[i] = *task
			return nil
		}
	}
	return fmt.Errorf("task no se ha encontrado")
}
