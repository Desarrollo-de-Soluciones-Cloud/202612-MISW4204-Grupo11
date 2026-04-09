package tasks

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type fakeRepo struct {
	tasks       map[int]*task
	attachments []*Attachment
	nextID      int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{tasks: map[int]*task{}, attachments: []*Attachment{}, nextID: 1}
}

func (repo *fakeRepo) Create(task *task) error {
	task.ID = repo.nextID
	repo.nextID++
	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (r *fakeRepo) GetAll() ([]task, error) {
	list := make([]task, 0, len(r.tasks))
	for _, task := range r.tasks {
		list = append(list, *task)
	}
	return list, nil
}

func (r *fakeRepo) GetByID(id string) (*task, error) {
	var intID int
	if _, err := fmt.Sscanf(id, "%d", &intID); err != nil {
		return nil, errTaskNotFound
	}
	task, ok := r.tasks[intID]
	if !ok {
		return nil, errTaskNotFound
	}
	copied := *task
	return &copied, nil
}

func (r *fakeRepo) Update(task *task) error {
	if _, ok := r.tasks[task.ID]; !ok {
		return errTaskNotFound
	}
	copied := *task
	r.tasks[task.ID] = &copied
	return nil
}

func (r *fakeRepo) Delete(id string) error {
	var intID int
	if _, err := fmt.Sscanf(id, "%d", &intID); err != nil {
		return errTaskNotFound
	}
	if _, ok := r.tasks[intID]; !ok {
		return errTaskNotFound
	}
	delete(r.tasks, intID)
	return nil
}

func (r *fakeRepo) SaveAttachment(attachment *Attachment) error {
	r.attachments = append(r.attachments, attachment)
	return nil
}

func (r *fakeRepo) UpdateStatus(task *task) error {
	return r.Update(task)
}

var errTaskNotFound = &taskNotFoundError{}

type taskNotFoundError struct{}

func (e *taskNotFoundError) Error() string { return "task not found" }

func TestTaskService_Create_ValidatesTitle(t *testing.T) {
	s := NewTaskService(newFakeRepo())
	err := s.Create(&task{Title: "", Status: StatusOpen}, &User{})
	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("wanted title validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesStatus(t *testing.T) {
	s := NewTaskService(newFakeRepo())
	err := s.Create(&task{Title: "task", Status: ""}, &User{})
	if err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("wanted invalid status error, got %v", err)
	}
}

func TestTaskService_Create_RejectsMoreThan22Hours(t *testing.T) {
	user := &User{Vinculations: []Vinculation{{Role: "assistant_graduated", Tasks: []task{{Week: 2, TimeInvested: 20}}}}}
	s := NewTaskService(newFakeRepo())
	err := s.Create(&task{Title: "task", Status: StatusOpen, Week: 2, TimeInvested: 3}, user)
	if err == nil || !strings.Contains(err.Error(), "No se pueden registrar mas horas") {
		t.Fatalf("wanted hours limit error, got %v", err)
	}
}

func TestTaskService_Create_SetsIDAndTimeRegistered(t *testing.T) {
	repo := newFakeRepo()
	s := NewTaskService(repo)
	taskItem := &task{Title: "task", Status: StatusOpen, Week: 1, TimeInvested: 4}
	err := s.Create(taskItem, &User{})
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
		t.Fatalf("expected task in repository after create")
	}
}

func TestTaskService_Delete_ReturnsErrorWhenOldOpenTask(t *testing.T) {
	repo := newFakeRepo()
	oldTask := &task{ID: 1, Status: StatusOpen, TimeRegistered: time.Now().Add(-8 * 24 * time.Hour)}
	repo.tasks[oldTask.ID] = oldTask
	s := NewTaskService(repo)
	err := s.Delete("1")
	if err == nil || !strings.Contains(err.Error(), "ya han pasado 7 días") {
		t.Fatalf("expected delete age validation error, got %v", err)
	}
}

func TestTaskService_Delete_DeletesWhenNotRestricted(t *testing.T) {
	repo := newFakeRepo()
	existing := &task{ID: 1, Status: StatusInDevelopment, TimeRegistered: time.Now().Add(-8 * 24 * time.Hour)}
	repo.tasks[existing.ID] = existing
	s := NewTaskService(repo)
	err := s.Delete("1")
	if err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
	if _, ok := repo.tasks[existing.ID]; ok {
		t.Fatal("expected task removed from repo")
	}
}

func TestTaskService_Update_ReturnsErrorWhenOldOpenTask(t *testing.T) {
	repo := newFakeRepo()
	oldTask := &task{ID: 1, Status: StatusOpen, TimeRegistered: time.Now().Add(-8 * 24 * time.Hour)}
	repo.tasks[oldTask.ID] = oldTask
	s := NewTaskService(repo)
	err := s.Update(oldTask)
	if err == nil || !strings.Contains(err.Error(), "7 days have passed by") {
		t.Fatalf("expected update age validation error, got %v", err)
	}
}

func TestTaskService_UpdateStatus_ReturnsErrorWhenOldOpenTask(t *testing.T) {
	repo := newFakeRepo()
	oldTask := &task{ID: 1, Status: StatusOpen, TimeRegistered: time.Now().Add(-8 * 24 * time.Hour)}
	repo.tasks[oldTask.ID] = oldTask
	s := NewTaskService(repo)
	err := s.UpdateStatus(oldTask)
	if err == nil || !strings.Contains(err.Error(), "7 days have passed by") {
		t.Fatalf("expected update status age validation error, got %v", err)
	}
}

func TestTaskService_UploadAttachment_ReturnsErrorWhenTaskMissing(t *testing.T) {
	s := NewTaskService(newFakeRepo())
	_, err := s.UploadAttachment("missing", &multipart.FileHeader{})
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Fatalf("expected task not found error, got %v", err)
	}
}

func TestTaskService_UploadAttachment_SavesFileAndAttachmentMetadata(t *testing.T) {
	tmpRepo := newFakeRepo()
	stored := &task{ID: 1, Status: StatusOpen, TimeRegistered: time.Now()}
	tmpRepo.tasks[stored.ID] = stored
	s := NewTaskService(tmpRepo)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")
	defer os.RemoveAll("./uploads")

	attachment, err := s.UploadAttachment("1", fileHeader)
	if err != nil {
		t.Fatalf("expected upload success, got %v", err)
	}
	if attachment.TaskID != "1" {
		t.Fatalf("expected attachment task id %q, got %q", "1", attachment.TaskID)
	}
	if attachment.FileName != "example.txt" {
		t.Fatalf("expected attachment filename example.txt, got %q", attachment.FileName)
	}
	if _, err := os.Stat(attachment.StoragePath); err != nil {
		t.Fatalf("expected saved file at %q, got error %v", attachment.StoragePath, err)
	}
	if len(tmpRepo.attachments) != 1 {
		t.Fatalf("expected repository to store one attachment, got %d", len(tmpRepo.attachments))
	}
}

func createMultipartFileHeader(t *testing.T, fieldName, filename, content string) *multipart.FileHeader {
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
	return req.MultipartForm.File[fieldName][0]
}
