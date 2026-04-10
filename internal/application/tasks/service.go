package tasks

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type TaskService struct {
	repo ports.TaskRepository
}

func NewTaskService(repo ports.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(task *domain.Task) error {
	if strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if strings.TrimSpace(task.Description) == "" {
		return fmt.Errorf("description is required")
	}

	if strings.TrimSpace(string(task.Status)) == "" {
		return fmt.Errorf("status is required")
	}

	if task.Week <= 0 {
		return fmt.Errorf("week must be greater than 0")
	}

	if task.TimeInvested <= 0 {
		return fmt.Errorf("time invested must be greater than 0")
	}

	if task.TimeInvested > 22 {
		return fmt.Errorf("no se pueden registrar más de 22 horas en una sola tarea")
	}

	if task.TimeRegistered.IsZero() {
		task.TimeRegistered = time.Now()
	}

	return s.repo.Create(task)
}

func (s *TaskService) GetAll() ([]domain.Task, error) {
	return s.repo.GetAll()
}

func (s *TaskService) GetByID(id string) (*domain.Task, error) {
	return s.repo.GetByID(id)
}

func (s *TaskService) Delete(taskID string) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}

	if isPast7Days(task.TimeRegistered) && isOpenStatus(task.Status) {
		return fmt.Errorf("ya han pasado 7 días, no se puede borrar")
	}

	return s.repo.Delete(taskID)
}

func (s *TaskService) Update(task *domain.Task) error {
	if task.ID <= 0 {
		return fmt.Errorf("invalid task id")
	}

	existingTask, err := s.repo.GetByID(strconv.Itoa(task.ID))
	if err != nil {
		return err
	}

	if isPast7Days(existingTask.TimeRegistered) {
		return fmt.Errorf("7 days have passed by, please create a new task")
	}

	if strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if strings.TrimSpace(task.Description) == "" {
		return fmt.Errorf("description is required")
	}

	if strings.TrimSpace(string(task.Status)) == "" {
		return fmt.Errorf("status is required")
	}

	if task.Week <= 0 {
		return fmt.Errorf("week must be greater than 0")
	}

	if task.TimeInvested <= 0 {
		return fmt.Errorf("time invested must be greater than 0")
	}

	if task.TimeInvested > 22 {
		return fmt.Errorf("no se pueden registrar más de 22 horas en una sola tarea")
	}

	task.TimeRegistered = existingTask.TimeRegistered

	return s.repo.Update(task)
}

func (s *TaskService) UpdateStatus(task *domain.Task) error {
	if task.ID <= 0 {
		return fmt.Errorf("invalid task id")
	}

	existingTask, err := s.repo.GetByID(strconv.Itoa(task.ID))
	if err != nil {
		return err
	}

	if isPast7Days(existingTask.TimeRegistered) {
		return fmt.Errorf("7 days have passed by, please create a new task")
	}

	if strings.TrimSpace(string(task.Status)) == "" {
		return fmt.Errorf("status is required")
	}

	existingTask.Status = task.Status

	return s.repo.UpdateStatus(existingTask)
}

func (s *TaskService) UploadAttachment(taskID string, file *multipart.FileHeader) (*domain.Attachment, error) {
	_, err := s.repo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}

	attachmentID := uuid.NewString()
	filePath := fmt.Sprintf("./uploads/%s_%s", attachmentID, file.Filename)

	if err := saveFile(file, filePath); err != nil {
		return nil, fmt.Errorf("could not save file: %w", err)
	}

	attachment := &domain.Attachment{
		ID:          attachmentID,
		TaskID:      taskID,
		FileName:    file.Filename,
		StoragePath: filePath,
	}

	if err := s.repo.SaveAttachment(attachment); err != nil {
		return nil, err
	}

	return attachment, nil
}

func saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func isPast7Days(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	return time.Since(t) >= 7*24*time.Hour
}

func isOpenStatus(status domain.Status) bool {
	value := strings.ToUpper(strings.TrimSpace(string(status)))
	return value == "OPEN" || value == "ABIERTO"
}
