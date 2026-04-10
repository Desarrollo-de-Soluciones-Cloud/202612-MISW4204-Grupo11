package tasks

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type fakeRepo struct {
	tasks       map[int]*domain.Task
	attachments []*domain.Attachment
	nextID      int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tasks:       map[int]*domain.Task{},
		attachments: []*domain.Attachment{},
		nextID:      1,
	}
}

func (repo *fakeRepo) Create(task *domain.Task) error {
	task.ID = repo.nextID
	repo.nextID++

	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (repo *fakeRepo) GetAll() ([]domain.Task, error) {
	list := make([]domain.Task, 0, len(repo.tasks))
	for _, task := range repo.tasks {
		list = append(list, *task)
	}
	return list, nil
}

func (repo *fakeRepo) GetByID(id string) (*domain.Task, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errTaskNotFound
	}

	task, ok := repo.tasks[intID]
	if !ok {
		return nil, errTaskNotFound
	}

	copied := *task
	return &copied, nil
}

func (repo *fakeRepo) Update(task *domain.Task) error {
	if _, ok := repo.tasks[task.ID]; !ok {
		return errTaskNotFound
	}

	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (repo *fakeRepo) Delete(id string) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return errTaskNotFound
	}

	if _, ok := repo.tasks[intID]; !ok {
		return errTaskNotFound
	}

	delete(repo.tasks, intID)
	return nil
}

func (repo *fakeRepo) SaveAttachment(attachment *domain.Attachment) error {
	repo.attachments = append(repo.attachments, attachment)
	return nil
}

func (repo *fakeRepo) UpdateStatus(task *domain.Task) error {
	stored, ok := repo.tasks[task.ID]
	if !ok {
		return errTaskNotFound
	}

	stored.Status = task.Status
	return nil
}

var errTaskNotFound = errors.New("task not found")

func TestTaskService_Create_ValidatesTitle(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	err := s.Create(&domain.Task{
		Title:        "",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 2,
	})

	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("wanted title validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesDescription(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	err := s.Create(&domain.Task{
		Title:        "task",
		Description:  "",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 2,
	})

	if err == nil || !strings.Contains(err.Error(), "description is required") {
		t.Fatalf("wanted description validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesStatus(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	err := s.Create(&domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status(""),
		Week:         1,
		TimeInvested: 2,
	})

	if err == nil || !strings.Contains(err.Error(), "status is required") {
		t.Fatalf("wanted status validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesWeek(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	err := s.Create(&domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         0,
		TimeInvested: 2,
	})

	if err == nil || !strings.Contains(err.Error(), "week must be greater than 0") {
		t.Fatalf("wanted week validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesTimeInvested(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	err := s.Create(&domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 0,
	})

	if err == nil || !strings.Contains(err.Error(), "time invested must be greater than 0") {
		t.Fatalf("wanted time invested validation error, got %v", err)
	}
}

func TestTaskService_Create_RejectsMoreThan22Hours(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	err := s.Create(&domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 23,
	})

	if err == nil || !strings.Contains(err.Error(), "22 horas") {
		t.Fatalf("wanted hours limit error, got %v", err)
	}
}

func TestTaskService_Create_SetsIDAndTimeRegistered(t *testing.T) {
	repo := newFakeRepo()
	s := NewTaskService(repo)

	taskItem := &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 4,
	}

	err := s.Create(taskItem)
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}

	if taskItem.ID == 0 {
		t.Fatal("expected generated ID")
	}

	if taskItem.TimeRegistered.IsZero() {
		t.Fatal("expected TimeRegistered to be set")
	}

	if _, ok := repo.tasks[taskItem.ID]; !ok {
		t.Fatal("expected task to be stored in repo")
	}
}

func TestTaskService_Delete_ReturnsErrorWhenOldOpenTask(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "task",
		Description:    "desc",
		Status:         domain.Status("OPEN"),
		Week:           1,
		TimeInvested:   3,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := NewTaskService(repo)

	err := s.Delete("1")
	if err == nil || !strings.Contains(err.Error(), "ya han pasado 7 días") {
		t.Fatalf("expected delete age validation error, got %v", err)
	}
}

func TestTaskService_Delete_DeletesWhenNotRestricted(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "task",
		Description:    "desc",
		Status:         domain.Status("IN_PROGRESS"),
		Week:           1,
		TimeInvested:   3,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := NewTaskService(repo)

	err := s.Delete("1")
	if err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}

	if _, ok := repo.tasks[1]; ok {
		t.Fatal("expected task removed from repo")
	}
}

func TestTaskService_Update_ReturnsErrorWhenTaskIsOlderThan7Days(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "old",
		Description:    "desc",
		Status:         domain.Status("OPEN"),
		Week:           1,
		TimeInvested:   3,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := NewTaskService(repo)

	err := s.Update(&domain.Task{
		ID:           1,
		Title:        "updated",
		Description:  "new desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 4,
	})

	if err == nil || !strings.Contains(err.Error(), "7 days have passed by") {
		t.Fatalf("expected update age validation error, got %v", err)
	}
}

func TestTaskService_Update_UpdatesWhenValid(t *testing.T) {
	repo := newFakeRepo()
	originalTime := time.Now().Add(-2 * 24 * time.Hour)

	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "old",
		Description:    "desc",
		Status:         domain.Status("OPEN"),
		Week:           1,
		TimeInvested:   3,
		TimeRegistered: originalTime,
	}

	s := NewTaskService(repo)

	err := s.Update(&domain.Task{
		ID:           1,
		Title:        "updated",
		Description:  "new desc",
		Status:       domain.Status("DONE"),
		Week:         2,
		TimeInvested: 5,
	})
	if err != nil {
		t.Fatalf("expected update success, got %v", err)
	}

	stored := repo.tasks[1]
	if stored.Title != "updated" {
		t.Fatalf("expected updated title, got %q", stored.Title)
	}
	if !stored.TimeRegistered.Equal(originalTime) {
		t.Fatalf("expected original TimeRegistered to be preserved")
	}
}

func TestTaskService_UpdateStatus_ReturnsErrorWhenTaskIsOlderThan7Days(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "task",
		Description:    "desc",
		Status:         domain.Status("OPEN"),
		Week:           1,
		TimeInvested:   2,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := NewTaskService(repo)

	err := s.UpdateStatus(&domain.Task{
		ID:     1,
		Status: domain.Status("DONE"),
	})

	if err == nil || !strings.Contains(err.Error(), "7 days have passed by") {
		t.Fatalf("expected update status age validation error, got %v", err)
	}
}

func TestTaskService_UpdateStatus_UpdatesWhenValid(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "task",
		Description:    "desc",
		Status:         domain.Status("OPEN"),
		Week:           1,
		TimeInvested:   2,
		TimeRegistered: time.Now().Add(-2 * 24 * time.Hour),
	}

	s := NewTaskService(repo)

	err := s.UpdateStatus(&domain.Task{
		ID:     1,
		Status: domain.Status("DONE"),
	})
	if err != nil {
		t.Fatalf("expected update status success, got %v", err)
	}

	if repo.tasks[1].Status != domain.Status("DONE") {
		t.Fatalf("expected status DONE, got %v", repo.tasks[1].Status)
	}
}

func TestTaskService_UploadAttachment_ReturnsErrorWhenTaskMissing(t *testing.T) {
	s := NewTaskService(newFakeRepo())

	_, err := s.UploadAttachment("999", &multipart.FileHeader{})
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Fatalf("expected task not found error, got %v", err)
	}
}

func TestTaskService_UploadAttachment_SavesFileAndAttachmentMetadata(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID:             1,
		Title:          "task",
		Description:    "desc",
		Status:         domain.Status("OPEN"),
		Week:           1,
		TimeInvested:   2,
		TimeRegistered: time.Now(),
	}

	s := NewTaskService(repo)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")
	defer os.RemoveAll("./uploads")

	attachment, err := s.UploadAttachment("1", fileHeader)
	if err != nil {
		t.Fatalf("expected upload success, got %v", err)
	}

	if attachment.TaskID != 1 {
		t.Fatalf("expected attachment task id %q, got %q", "1", attachment.TaskID)
	}

	if attachment.FileName != "example.txt" {
		t.Fatalf("expected attachment filename example.txt, got %q", attachment.FileName)
	}

	if _, err := os.Stat(attachment.StoragePath); err != nil {
		t.Fatalf("expected saved file at %q, got error %v", attachment.StoragePath, err)
	}

	if len(repo.attachments) != 1 {
		t.Fatalf("expected repository to store one attachment, got %d", len(repo.attachments))
	}

	if attachment.ID == 0 {
		t.Fatal("expected generated attachment ID")
	}
}

func createMultipartFileHeader(t *testing.T, fieldName, filename, content string) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatal(err)
	}

	files := req.MultipartForm.File[fieldName]
	if len(files) == 0 {
		t.Fatal("expected multipart file header")
	}

	return files[0]
}
