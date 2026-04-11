package tasks

import (
	"bytes"
	"context"
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
	tasks                map[int]*domain.Task
	attachments          []*domain.Attachment
	nextID               int
	assignmentUsers      map[int]int64
	assignmentProfessors map[int]int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tasks:                map[int]*domain.Task{},
		attachments:          []*domain.Attachment{},
		nextID:               1,
		assignmentUsers:      map[int]int64{1: 1},
		assignmentProfessors: map[int]int64{1: 10},
	}
}

func (repo *fakeRepo) Create(task *domain.Task) error {
	task.ID = repo.nextID
	repo.nextID++

	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (repo *fakeRepo) ListAll(ctx context.Context) ([]domain.Task, error) {
	list := make([]domain.Task, 0, len(repo.tasks))
	for _, task := range repo.tasks {
		list = append(list, *task)
	}
	return list, nil
}

func (repo *fakeRepo) ListByUser(ctx context.Context, userID int64) ([]domain.Task, error) {
	var list []domain.Task
	for _, task := range repo.tasks {
		owner, ok := repo.assignmentUsers[task.AssignmentId]
		if ok && owner == userID {
			list = append(list, *task)
		}
	}
	return list, nil
}

func (repo *fakeRepo) ListByProfessorID(ctx context.Context, professorID int64) ([]domain.Task, error) {
	var list []domain.Task
	for _, task := range repo.tasks {
		prof, ok := repo.assignmentProfessors[task.AssignmentId]
		if ok && prof == professorID {
			list = append(list, *task)
		}
	}
	return list, nil
}

func (repo *fakeRepo) GetByID(id string) (*domain.Task, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errLegacyTaskNotFound
	}

	task, ok := repo.tasks[intID]
	if !ok {
		return nil, errLegacyTaskNotFound
	}

	copied := *task
	return &copied, nil
}

func (repo *fakeRepo) GetByIDForUser(ctx context.Context, id string, userID int64) (*domain.Task, error) {
	task, err := repo.GetByID(id)
	if err != nil {
		return nil, domain.ErrTaskNotFound
	}
	owner, ok := repo.assignmentUsers[task.AssignmentId]
	if !ok || owner != userID {
		return nil, domain.ErrTaskNotFound
	}
	return task, nil
}

func (repo *fakeRepo) Update(task *domain.Task) error {
	if _, ok := repo.tasks[task.ID]; !ok {
		return errLegacyTaskNotFound
	}

	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (repo *fakeRepo) Delete(id string) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return errLegacyTaskNotFound
	}

	if _, ok := repo.tasks[intID]; !ok {
		return errLegacyTaskNotFound
	}

	delete(repo.tasks, intID)
	return nil
}

func (repo *fakeRepo) SaveAttachment(attachment *domain.Attachment) error {
	attachment.ID = repo.nextID
	repo.nextID++
	repo.attachments = append(repo.attachments, attachment)
	return nil
}

func (repo *fakeRepo) UpdateStatus(task *domain.Task) error {
	stored, ok := repo.tasks[task.ID]
	if !ok {
		return errLegacyTaskNotFound
	}

	stored.Status = task.Status
	return nil
}

var errLegacyTaskNotFound = errors.New("task not found")

type fakeAssignmentRepo struct {
	byID map[int64]*domain.Assignment
}

func newFakeAssignmentRepo() *fakeAssignmentRepo {
	return &fakeAssignmentRepo{byID: map[int64]*domain.Assignment{
		1: {ID: 1, UserID: 1, ProfessorID: 10},
	}}
}

func (f *fakeAssignmentRepo) Create(_ context.Context, a *domain.Assignment) error {
	f.byID[a.ID] = a
	return nil
}

func (f *fakeAssignmentRepo) FindByID(_ context.Context, id int64) (*domain.Assignment, error) {
	a, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrVinculacionNoEncontrada
	}
	return a, nil
}

func (f *fakeAssignmentRepo) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}

func (f *fakeAssignmentRepo) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}

func (f *fakeAssignmentRepo) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return false, nil
}

func (f *fakeAssignmentRepo) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}

func (f *fakeAssignmentRepo) FindByProfessorWithUser(_ context.Context, _ int64) ([]domain.AssignmentWithUser, error) {
	return nil, nil
}

func (f *fakeAssignmentRepo) Update(_ context.Context, _ *domain.Assignment) error {
	return nil
}

func (f *fakeAssignmentRepo) ListAll(_ context.Context) ([]domain.Assignment, error) {
	out := make([]domain.Assignment, 0, len(f.byID))
	for _, a := range f.byID {
		if a != nil {
			out = append(out, *a)
		}
	}
	return out, nil
}

func newTaskServiceForTest(repo *fakeRepo) *TaskService {
	return NewTaskService(repo, newFakeAssignmentRepo())
}

func TestTaskService_Create_ValidatesTitle(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	err := s.Create(context.Background(), &domain.Task{
		Title:        "",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 2,
		AssignmentId: 1,
	}, 1)

	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("wanted title validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesDescription(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	err := s.Create(context.Background(), &domain.Task{
		Title:        "task",
		Description:  "",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 2,
		AssignmentId: 1,
	}, 1)

	if err == nil || !strings.Contains(err.Error(), "description is required") {
		t.Fatalf("wanted description validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesStatus(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	err := s.Create(context.Background(), &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status(""),
		Week:         1,
		TimeInvested: 2,
		AssignmentId: 1,
	}, 1)

	if err == nil || !strings.Contains(err.Error(), "status is required") {
		t.Fatalf("wanted status validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesWeek(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	err := s.Create(context.Background(), &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         0,
		TimeInvested: 2,
		AssignmentId: 1,
	}, 1)

	if err == nil || !strings.Contains(err.Error(), "week must be greater than 0") {
		t.Fatalf("wanted week validation error, got %v", err)
	}
}

func TestTaskService_Create_ValidatesTimeInvested(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	err := s.Create(context.Background(), &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 0,
		AssignmentId: 1,
	}, 1)

	if err == nil || !strings.Contains(err.Error(), "time invested must be greater than 0") {
		t.Fatalf("wanted time invested validation error, got %v", err)
	}
}

func TestTaskService_Create_RejectsMoreThan22Hours(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	err := s.Create(context.Background(), &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 23,
		AssignmentId: 1,
	}, 1)

	if err == nil || !strings.Contains(err.Error(), "22 horas") {
		t.Fatalf("wanted hours limit error, got %v", err)
	}
}

func TestTaskService_Create_RejectsForeignAssignment(t *testing.T) {
	assignRepo := &fakeAssignmentRepo{byID: map[int64]*domain.Assignment{
		1: {ID: 1, UserID: 2},
	}}
	repo := newFakeRepo()
	s := NewTaskService(repo, assignRepo)

	err := s.Create(context.Background(), &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 2,
		AssignmentId: 1,
	}, 1)

	if !errors.Is(err, domain.ErrAssignmentNotOwned) {
		t.Fatalf("wanted ErrAssignmentNotOwned, got %v", err)
	}
}

func TestTaskService_Create_SetsIDAndTimeRegistered(t *testing.T) {
	repo := newFakeRepo()
	s := NewTaskService(repo, newFakeAssignmentRepo())

	taskItem := &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 4,
		AssignmentId: 1,
	}

	err := s.Create(context.Background(), taskItem, 1)
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
		AssignmentId:   1,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := newTaskServiceForTest(repo)

	err := s.Delete(context.Background(), "1", 1)
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
		AssignmentId:   1,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := newTaskServiceForTest(repo)

	err := s.Delete(context.Background(), "1", 1)
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
		AssignmentId:   1,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := newTaskServiceForTest(repo)

	err := s.Update(context.Background(), &domain.Task{
		ID:           1,
		Title:        "updated",
		Description:  "new desc",
		Status:       domain.Status("OPEN"),
		Week:         1,
		TimeInvested: 4,
		AssignmentId: 1,
	}, 1)

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
		AssignmentId:   1,
		TimeRegistered: originalTime,
	}

	s := newTaskServiceForTest(repo)

	err := s.Update(context.Background(), &domain.Task{
		ID:           1,
		Title:        "updated",
		Description:  "new desc",
		Status:       domain.Status("DONE"),
		Week:         2,
		TimeInvested: 5,
		AssignmentId: 1,
	}, 1)
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
		AssignmentId:   1,
		TimeRegistered: time.Now().Add(-8 * 24 * time.Hour),
	}

	s := newTaskServiceForTest(repo)

	err := s.UpdateStatus(context.Background(), &domain.Task{
		ID:     1,
		Status: domain.Status("DONE"),
	}, 1)

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
		AssignmentId:   1,
		TimeRegistered: time.Now().Add(-2 * 24 * time.Hour),
	}

	s := newTaskServiceForTest(repo)

	err := s.UpdateStatus(context.Background(), &domain.Task{
		ID:     1,
		Status: domain.Status("DONE"),
	}, 1)
	if err != nil {
		t.Fatalf("expected update status success, got %v", err)
	}

	if repo.tasks[1].Status != domain.Status("DONE") {
		t.Fatalf("expected status DONE, got %v", repo.tasks[1].Status)
	}
}

func TestTaskService_UploadAttachment_ReturnsErrorWhenTaskMissing(t *testing.T) {
	s := newTaskServiceForTest(newFakeRepo())

	_, err := s.UploadAttachment(context.Background(), "999", 1, &multipart.FileHeader{})
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
		AssignmentId:   1,
		TimeRegistered: time.Now(),
	}

	s := newTaskServiceForTest(repo)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")
	defer os.RemoveAll("./uploads")

	attachment, err := s.UploadAttachment(context.Background(), "1", 1, fileHeader)
	if err != nil {
		t.Fatalf("expected upload success, got %v", err)
	}

	if attachment.TaskID != 1 {
		t.Fatalf("expected attachment task id 1, got %d", attachment.TaskID)
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
