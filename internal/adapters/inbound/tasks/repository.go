package tasks

import "fmt"

type Repository interface {
	Create(task *task) error
	GetAll() ([]task, error)
	GetByID(id string) (*task, error)
	Update(task *task) error
	Delete(id string) error
}

type TaskRepository struct {
	tasks []task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: []task{},
	}
}

func (r *TaskRepository) Create(task *task) error {
	r.tasks = append(r.tasks, *task)
	return nil
}

func (r *TaskRepository) GetAll() ([]task, error) {
	return r.tasks, nil
}

func (r *TaskRepository) GetByID(id string) (*task, error) {
	for i := range r.tasks {
		if r.tasks[i].ID == id {
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
	return fmt.Errorf("task not found")
}

func (r *TaskRepository) Delete(id string) error {
	for i := range r.tasks {
		if r.tasks[i].ID == id {
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task not found")
}
