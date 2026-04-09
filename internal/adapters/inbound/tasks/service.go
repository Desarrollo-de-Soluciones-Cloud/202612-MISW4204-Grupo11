package tasks

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

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
	task.TimeRegistered = time.Now()
	return s.repo.Create(task)
}

func (s *TaskService) GetAll() ([]task, error) {
	return s.repo.GetAll()
}

func (s *TaskService) Delete(taskID string) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}

	if time.Since(task.TimeRegistered) >= 7*24*time.Hour && task.Status == StatusOpen {
		return fmt.Errorf("ya han pasado 7 días, no se puede borrar.")
	}
	return s.repo.Delete(taskID)
}

func (s *TaskService) Update(task *task) error {
	if time.Since(task.TimeRegistered) >= 7*24*time.Hour && task.Status == StatusOpen {
		return fmt.Errorf("ya han pasado 7 días, no se puede modificar; crea una nueva tarea")
	}

	return s.repo.Update(task)
}

func (s *TaskService) UpdateStatus(task *task) error {
	if time.Since(task.TimeRegistered) >= 7*24*time.Hour && task.Status == StatusOpen {
		return fmt.Errorf("ya han pasado 7 días, no se puede modificar; crea una nueva tarea")
	}

	return s.repo.UpdateStatus(task)

}

func (s *TaskService) UploadAttachment(taskID string, file *multipart.FileHeader) (*Attachment, error) {
	// 1. verificar que la tarea exista
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}

	_ = task

	// 2. generar id del adjunto
	attachmentID := uuid.NewString()

	// 3. construir ruta
	filePath := fmt.Sprintf("./uploads/%s_%s", attachmentID, file.Filename)

	// 4. guardar archivo físico
	err = saveFile(file, filePath)
	if err != nil {
		return nil, fmt.Errorf("could not save file")
	}

	// 5. crear metadata
	attachment := &Attachment{
		ID:          attachmentID,
		TaskID:      taskID,
		FileName:    file.Filename,
		StoragePath: filePath,
	}

	// 6. guardar metadata en BD
	err = s.repo.SaveAttachment(attachment)
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

/*func limitOfTimeRecorded22(user *User, week int, newTaskHours int) bool {
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
}*/

func saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
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
