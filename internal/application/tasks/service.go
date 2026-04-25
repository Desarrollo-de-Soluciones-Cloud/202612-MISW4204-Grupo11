package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type TaskService struct {
	repo        ports.TaskRepository
	assignments domain.AssignmentRepository
	NowFunc     func() time.Time
}

const (
	ErrMaxTimeInvestedPerTask    = "no se pueden registrar más de 22 horas en una sola tarea"
	ErrTitleRequired             = "title is required"
	ErrDescriptionRequired       = "description is required"
	ErrStatusRequired            = "status is required"
	ErrAssignmentIDRequired      = "assignment_id is required"
	ErrTimeInvestedMustBeGreater = "time invested must be greater than 0"
	ErrInvalidTaskID             = "invalid task id"
)

type UpdateTaskInput struct {
	Title        *string
	Description  *string
	Status       *domain.Status
	TimeInvested *int
	Observations *string
}

func NewTaskService(repo ports.TaskRepository, assignments domain.AssignmentRepository) *TaskService {
	return &TaskService{repo: repo, assignments: assignments, NowFunc: time.Now}
}

func (s *TaskService) currentWeekStart() time.Time {
	return domain.WeekStartFor(s.NowFunc())
}

func (s *TaskService) isCurrentWeek(weekStart time.Time) bool {
	return weekStart.Equal(s.currentWeekStart())
}

func (s *TaskService) Create(ctx context.Context, task *domain.Task, currentUserID int64) error {
	if strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf(ErrTitleRequired)
	}

	if strings.TrimSpace(task.Description) == "" {
		return fmt.Errorf(ErrDescriptionRequired)
	}

	if strings.TrimSpace(string(task.Status)) == "" {
		return fmt.Errorf(ErrStatusRequired)
	}

	if err := domain.ValidateWeekStart(task.WeekStart); err != nil {
		return err
	}

	currentMonday := s.currentWeekStart()
	if task.WeekStart.After(currentMonday) {
		return domain.ErrSemanaFutura
	}

	if task.WeekStart.Before(currentMonday) {
		task.IsLate = true
	}

	if task.TimeInvested <= 0 {
		return fmt.Errorf(ErrTimeInvestedMustBeGreater)
	}

	if task.AssignmentId <= 0 {
		return fmt.Errorf(ErrAssignmentIDRequired)
	}

	assignment, err := s.assignments.FindByID(ctx, int64(task.AssignmentId))
	if err != nil {
		return err
	}
	if assignment.UserID != currentUserID {
		return domain.ErrAssignmentNotOwned
	}

	if task.TimeInvested > 22 {
		return fmt.Errorf(ErrMaxTimeInvestedPerTask)
	}

	if task.TimeRegistered.IsZero() {
		task.TimeRegistered = s.NowFunc()
	}

	return s.repo.Create(task)
}

func (s *TaskService) ListForUser(ctx context.Context, userID int64) ([]domain.Task, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *TaskService) ListForProfessor(ctx context.Context, professorID int64) ([]domain.Task, error) {
	return s.repo.ListByProfessorID(ctx, professorID)
}

func (s *TaskService) ListAllForAdmin(ctx context.Context) ([]domain.Task, error) {
	return s.repo.ListAll(ctx)
}

func (s *TaskService) GetByIDForUser(ctx context.Context, id string, userID int64) (*domain.Task, error) {
	return s.repo.GetByIDForUser(ctx, id, userID)
}

func (s *TaskService) Delete(ctx context.Context, taskID string, userID int64) error {
	task, err := s.repo.GetByIDForUser(ctx, taskID, userID)
	if err != nil {
		return err
	}

	if task.IsLate {
		return domain.ErrReporteTardioNoEliminable
	}

	if !s.isCurrentWeek(task.WeekStart) {
		return domain.ErrEliminacionFueraDeSemana
	}

	return s.repo.Delete(taskID)
}

func (s *TaskService) Update(ctx context.Context, task *domain.Task, userID int64) error {
	if task.ID <= 0 {
		return fmt.Errorf(ErrInvalidTaskID)
	}

	existingTask, err := s.repo.GetByIDForUser(ctx, strconv.Itoa(task.ID), userID)
	if err != nil {
		return err
	}

	if existingTask.IsLate {
		return domain.ErrReporteTardioInmutable
	}

	if !s.isCurrentWeek(existingTask.WeekStart) {
		return domain.ErrModificacionFueraDeSemana
	}

	if strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf(ErrTitleRequired)
	}

	if strings.TrimSpace(task.Description) == "" {
		return fmt.Errorf(ErrDescriptionRequired)
	}

	if strings.TrimSpace(string(task.Status)) == "" {
		return fmt.Errorf(ErrStatusRequired)
	}

	if task.TimeInvested <= 0 {
		return fmt.Errorf(ErrTimeInvestedMustBeGreater)
	}

	if task.TimeInvested > 22 {
		return fmt.Errorf(ErrMaxTimeInvestedPerTask)
	}

	task.TimeRegistered = existingTask.TimeRegistered
	task.IsLate = existingTask.IsLate
	task.WeekStart = existingTask.WeekStart

	return s.repo.Update(task)
}

func (s *TaskService) PartialUpdate(ctx context.Context, id string, userID int64, input UpdateTaskInput) (*domain.Task, error) {
	task, err := s.repo.GetByIDForUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if task.IsLate {
		return nil, domain.ErrReporteTardioInmutable
	}

	if !s.isCurrentWeek(task.WeekStart) {
		return nil, domain.ErrModificacionFueraDeSemana
	}

	if input.Title != nil {
		task.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		task.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		task.Status = *input.Status
	}
	if input.TimeInvested != nil {
		task.TimeInvested = *input.TimeInvested
	}
	if input.Observations != nil {
		task.Observations = *input.Observations
	}

	if strings.TrimSpace(task.Title) == "" {
		return nil, fmt.Errorf(ErrTitleRequired)
	}
	if strings.TrimSpace(task.Description) == "" {
		return nil, fmt.Errorf(ErrDescriptionRequired)
	}
	if strings.TrimSpace(string(task.Status)) == "" {
		return nil, fmt.Errorf(ErrStatusRequired)
	}
	if task.TimeInvested <= 0 {
		return nil, fmt.Errorf(ErrTimeInvestedMustBeGreater)
	}
	if task.TimeInvested > 22 {
		return nil, fmt.Errorf(ErrMaxTimeInvestedPerTask)
	}

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) UpdateStatus(ctx context.Context, task *domain.Task, userID int64) error {
	if task.ID <= 0 {
		return fmt.Errorf(ErrInvalidTaskID)
	}

	existingTask, err := s.repo.GetByIDForUser(ctx, strconv.Itoa(task.ID), userID)
	if err != nil {
		return err
	}

	if existingTask.IsLate {
		return domain.ErrReporteTardioInmutable
	}

	if !s.isCurrentWeek(existingTask.WeekStart) {
		return domain.ErrModificacionFueraDeSemana
	}

	if strings.TrimSpace(string(task.Status)) == "" {
		return fmt.Errorf(ErrStatusRequired)
	}

	existingTask.Status = task.Status

	return s.repo.UpdateStatus(existingTask)
}

func (s *TaskService) UploadAttachment(ctx context.Context, taskID string, userID int64, file *multipart.FileHeader) (*domain.Attachment, error) {
	_, err := s.repo.GetByIDForUser(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	taskIDInt, err := strconv.Atoi(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id")
	}

	uniqueName := fmt.Sprintf("%d_%s", s.NowFunc().UnixNano(), file.Filename)
	filePath := filepath.Join("./uploads", uniqueName)

	if err := saveFile(file, filePath); err != nil {
		return nil, fmt.Errorf("could not save file: %w", err)
	}

	contentType := file.Header.Get("Content-Type")

	attachment := &domain.Attachment{
		TaskID:      taskIDInt,
		FileName:    file.Filename,
		ContentType: contentType,
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
