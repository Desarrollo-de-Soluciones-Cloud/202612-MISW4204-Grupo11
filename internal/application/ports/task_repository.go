package ports

import (
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type TaskRepository interface {
	Create(task *domain.Task) error
	GetAll() ([]domain.Task, error)
	GetByID(id string) (*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id string) error
	SaveAttachment(attachment *domain.Attachment) error
	UpdateStatus(task *domain.Task) error
}
