package ports

import (
	"context"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type TaskRepository interface {
	Create(task *domain.Task) error
	ListAll(ctx context.Context) ([]domain.Task, error)
	ListByUser(ctx context.Context, userID int64) ([]domain.Task, error)
	ListByProfessorID(ctx context.Context, professorID int64) ([]domain.Task, error)
	GetByID(id string) (*domain.Task, error)
	GetByIDForUser(ctx context.Context, id string, userID int64) (*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id string) error
	SaveAttachment(attachment *domain.Attachment) error
	UpdateStatus(task *domain.Task) error
	GetAttachments(ctx context.Context, taskID int) ([]domain.Attachment, error)
	ListByAssignmentAndWeek(ctx context.Context, assignmentID int64, weekStart time.Time) ([]domain.Task, error)
}
