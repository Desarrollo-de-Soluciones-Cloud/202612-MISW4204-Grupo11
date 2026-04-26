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

func (repo *fakeRepo) ListAll(_ context.Context) ([]domain.Task, error) {
	list := make([]domain.Task, 0, len(repo.tasks))
	for _, task := range repo.tasks {
		list = append(list, *task)
	}
	return list, nil
}

func (repo *fakeRepo) ListByUser(_ context.Context, userID int64) ([]domain.Task, error) {
	var list []domain.Task
	for _, task := range repo.tasks {
		owner, ok := repo.assignmentUsers[task.AssignmentId]
		if ok && owner == userID {
			list = append(list, *task)
		}
	}
	return list, nil
}

func (repo *fakeRepo) ListByProfessorID(_ context.Context, professorID int64) ([]domain.Task, error) {
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

func (repo *fakeRepo) GetByIDForUser(_ context.Context, id string, userID int64) (*domain.Task, error) {
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

func (repo *fakeRepo) ListByAssignmentAndWeek(_ context.Context, _ int64, _ time.Time) ([]domain.Task, error) {
	return nil, nil
}

func (repo *fakeRepo) ListByAssignment(_ context.Context, assignmentID int64) ([]domain.Task, error) {
	var result []domain.Task
	for _, task := range repo.tasks {
		if task.AssignmentId == int(assignmentID) {
			result = append(result, *task)
		}
	}
	return result, nil
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
func (f *fakeAssignmentRepo) Update(_ context.Context, _ *domain.Assignment) error { return nil }
func (f *fakeAssignmentRepo) ListAll(_ context.Context) ([]domain.Assignment, error) {
	out := make([]domain.Assignment, 0, len(f.byID))
	for _, a := range f.byID {
		if a != nil {
			out = append(out, *a)
		}
	}
	return out, nil
}

// fixedNow is a Wednesday so tests have a clear current week.
var fixedNow = time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC) // Wednesday 2026-04-08
var thisMonday = time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
var lastMonday = time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
var nextMonday = time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

const testUserID int64 = 1

func newTestService(repo *fakeRepo) *TaskService {
	svc := NewTaskService(repo, newFakeAssignmentRepo())
	svc.NowFunc = func() time.Time { return fixedNow }
	return svc
}

func validTask(weekStart time.Time) *domain.Task {
	return &domain.Task{
		Title:        "task",
		Description:  "desc",
		Status:       domain.StatusOpen,
		WeekStart:    weekStart,
		TimeInvested: 4,
		AssignmentId: 1,
	}
}

// --- Create tests ---

func TestCreate_CurrentWeek_NotLate(t *testing.T) {
	repo := newFakeRepo()
	s := newTestService(repo)

	task := validTask(thisMonday)
	if err := s.Create(context.Background(), task, testUserID); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if task.IsLate {
		t.Fatal("expected IsLate=false for current week")
	}
	if task.ID == 0 {
		t.Fatal("expected generated ID")
	}
	if task.TimeRegistered.IsZero() {
		t.Fatal("expected TimeRegistered to be set")
	}
}

func TestCreate_PastWeek_AutoLate(t *testing.T) {
	s := newTestService(newFakeRepo())

	task := validTask(lastMonday)
	if err := s.Create(context.Background(), task, testUserID); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !task.IsLate {
		t.Fatal("expected IsLate=true for past week")
	}
}

func TestCreate_FutureWeek_Rejected(t *testing.T) {
	s := newTestService(newFakeRepo())

	task := validTask(nextMonday)
	err := s.Create(context.Background(), task, testUserID)
	if !errors.Is(err, domain.ErrSemanaFutura) {
		t.Fatalf("expected ErrSemanaFutura, got %v", err)
	}
}

func TestCreate_NotMonday_Rejected(t *testing.T) {
	s := newTestService(newFakeRepo())

	tuesday := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)
	task := validTask(tuesday)
	err := s.Create(context.Background(), task, testUserID)
	if !errors.Is(err, domain.ErrSemanaInicioNoEsLunes) {
		t.Fatalf("expected ErrSemanaInicioNoEsLunes, got %v", err)
	}
}

func TestCreate_ValidatesTitle(t *testing.T) {
	s := newTestService(newFakeRepo())
	task := validTask(thisMonday)
	task.Title = ""
	err := s.Create(context.Background(), task, testUserID)
	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("wanted title validation error, got %v", err)
	}
}

func TestCreate_ValidatesDescription(t *testing.T) {
	s := newTestService(newFakeRepo())
	task := validTask(thisMonday)
	task.Description = ""
	err := s.Create(context.Background(), task, testUserID)
	if err == nil || !strings.Contains(err.Error(), "description is required") {
		t.Fatalf("wanted description validation error, got %v", err)
	}
}

func TestCreate_ValidatesStatus(t *testing.T) {
	s := newTestService(newFakeRepo())
	task := validTask(thisMonday)
	task.Status = ""
	err := s.Create(context.Background(), task, testUserID)
	if err == nil || !strings.Contains(err.Error(), "status is required") {
		t.Fatalf("wanted status validation error, got %v", err)
	}
}

func TestCreate_ValidatesTimeInvested(t *testing.T) {
	s := newTestService(newFakeRepo())
	task := validTask(thisMonday)
	task.TimeInvested = 0
	err := s.Create(context.Background(), task, testUserID)
	if err == nil || !strings.Contains(err.Error(), "time invested must be greater than 0") {
		t.Fatalf("wanted time invested validation error, got %v", err)
	}
}

func TestCreate_RejectsMoreThan22Hours(t *testing.T) {
	s := newTestService(newFakeRepo())
	task := validTask(thisMonday)
	task.TimeInvested = 23
	err := s.Create(context.Background(), task, testUserID)
	if err == nil || !strings.Contains(err.Error(), "22 horas") {
		t.Fatalf("wanted hours limit error, got %v", err)
	}
}

func TestCreate_RejectsForeignAssignment(t *testing.T) {
	assignRepo := &fakeAssignmentRepo{byID: map[int64]*domain.Assignment{
		1: {ID: 1, UserID: 2},
	}}
	repo := newFakeRepo()
	svc := NewTaskService(repo, assignRepo)
	svc.NowFunc = func() time.Time { return fixedNow }

	task := validTask(thisMonday)
	err := svc.Create(context.Background(), task, testUserID)
	if !errors.Is(err, domain.ErrAssignmentNotOwned) {
		t.Fatalf("wanted ErrAssignmentNotOwned, got %v", err)
	}
}

// --- Update tests ---

func TestUpdate_CurrentWeek_OK(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "old", Description: "desc", Status: domain.StatusOpen,
		WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow.Add(-24 * time.Hour),
	}
	s := newTestService(repo)

	err := s.Update(context.Background(), &domain.Task{
		ID: 1, Title: "updated", Description: "new desc",
		Status: domain.StatusInDevelopment, TimeInvested: 5, AssignmentId: 1,
	}, testUserID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if repo.tasks[1].Title != "updated" {
		t.Fatalf("expected title 'updated', got %q", repo.tasks[1].Title)
	}
	if repo.tasks[1].WeekStart != thisMonday {
		t.Fatal("expected WeekStart preserved")
	}
}

func TestUpdate_PastWeek_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "old", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow.Add(-10 * 24 * time.Hour),
	}
	s := newTestService(repo)

	err := s.Update(context.Background(), &domain.Task{
		ID: 1, Title: "updated", Description: "new desc",
		Status: domain.StatusOpen, TimeInvested: 4, AssignmentId: 1,
	}, testUserID)
	if !errors.Is(err, domain.ErrModificacionFueraDeSemana) {
		t.Fatalf("expected ErrModificacionFueraDeSemana, got %v", err)
	}
}

func TestUpdate_LateReport_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "late task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, IsLate: true, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow.Add(-24 * time.Hour),
	}
	s := newTestService(repo)

	err := s.Update(context.Background(), &domain.Task{
		ID: 1, Title: "updated", Description: "new desc",
		Status: domain.StatusOpen, TimeInvested: 4, AssignmentId: 1,
	}, testUserID)
	if !errors.Is(err, domain.ErrReporteTardioInmutable) {
		t.Fatalf("expected ErrReporteTardioInmutable, got %v", err)
	}
}

func TestUpdate_PreservesTimeRegistered(t *testing.T) {
	repo := newFakeRepo()
	originalTime := fixedNow.Add(-2 * 24 * time.Hour)
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "old", Description: "desc", Status: domain.StatusOpen,
		WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: originalTime,
	}
	s := newTestService(repo)

	err := s.Update(context.Background(), &domain.Task{
		ID: 1, Title: "updated", Description: "new desc",
		Status: domain.StatusFinalized, TimeInvested: 5, AssignmentId: 1,
	}, testUserID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !repo.tasks[1].TimeRegistered.Equal(originalTime) {
		t.Fatal("expected original TimeRegistered to be preserved")
	}
}

// --- Delete tests ---

func TestDelete_CurrentWeek_OK(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1, TimeRegistered: fixedNow,
	}
	s := newTestService(repo)

	if err := s.Delete(context.Background(), "1", testUserID); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if _, ok := repo.tasks[1]; ok {
		t.Fatal("expected task removed from repo")
	}
}

func TestDelete_PastWeek_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow.Add(-10 * 24 * time.Hour),
	}
	s := newTestService(repo)

	err := s.Delete(context.Background(), "1", testUserID)
	if !errors.Is(err, domain.ErrEliminacionFueraDeSemana) {
		t.Fatalf("expected ErrEliminacionFueraDeSemana, got %v", err)
	}
}

func TestDelete_LateReport_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, IsLate: true, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow,
	}
	s := newTestService(repo)

	err := s.Delete(context.Background(), "1", testUserID)
	if !errors.Is(err, domain.ErrReporteTardioNoEliminable) {
		t.Fatalf("expected ErrReporteTardioNoEliminable, got %v", err)
	}
}

// --- PartialUpdate tests ---

func TestPartialUpdate_CurrentWeek_OK(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1, TimeRegistered: fixedNow,
	}
	s := newTestService(repo)

	newTitle := "updated title"
	task, err := s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Title: &newTitle})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if task.Title != "updated title" {
		t.Fatalf("expected title 'updated title', got %q", task.Title)
	}
}

func TestPartialUpdate_LateReport_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, IsLate: true, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow,
	}
	s := newTestService(repo)

	newTitle := "updated"
	_, err := s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Title: &newTitle})
	if !errors.Is(err, domain.ErrReporteTardioInmutable) {
		t.Fatalf("expected ErrReporteTardioInmutable, got %v", err)
	}
}

func TestPartialUpdate_PastWeek_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, TimeInvested: 3, AssignmentId: 1,
		TimeRegistered: fixedNow.Add(-10 * 24 * time.Hour),
	}
	s := newTestService(repo)

	newTitle := "updated"
	_, err := s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Title: &newTitle})
	if !errors.Is(err, domain.ErrModificacionFueraDeSemana) {
		t.Fatalf("expected ErrModificacionFueraDeSemana, got %v", err)
	}
}

// --- UpdateStatus tests ---

func TestUpdateStatus_CurrentWeek_OK(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow,
	}
	s := newTestService(repo)

	err := s.UpdateStatus(context.Background(), &domain.Task{ID: 1, Status: domain.StatusFinalized}, testUserID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if repo.tasks[1].Status != domain.StatusFinalized {
		t.Fatalf("expected status %v, got %v", domain.StatusFinalized, repo.tasks[1].Status)
	}
}

func TestUpdateStatus_LateReport_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, IsLate: true, TimeInvested: 2, AssignmentId: 1,
		TimeRegistered: fixedNow,
	}
	s := newTestService(repo)

	err := s.UpdateStatus(context.Background(), &domain.Task{ID: 1, Status: domain.StatusFinalized}, testUserID)
	if !errors.Is(err, domain.ErrReporteTardioInmutable) {
		t.Fatalf("expected ErrReporteTardioInmutable, got %v", err)
	}
}

func TestUpdateStatus_PastWeek_Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: lastMonday, TimeInvested: 2, AssignmentId: 1,
		TimeRegistered: fixedNow.Add(-10 * 24 * time.Hour),
	}
	s := newTestService(repo)

	err := s.UpdateStatus(context.Background(), &domain.Task{ID: 1, Status: domain.StatusFinalized}, testUserID)
	if !errors.Is(err, domain.ErrModificacionFueraDeSemana) {
		t.Fatalf("expected ErrModificacionFueraDeSemana, got %v", err)
	}
}

// --- Attachment tests ---

func TestUploadAttachment_TaskNotFound(t *testing.T) {
	s := newTestService(newFakeRepo())

	_, err := s.UploadAttachment(context.Background(), "999", testUserID, &multipart.FileHeader{})
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Fatalf("expected task not found error, got %v", err)
	}
}

func TestUploadAttachment_SavesFileAndMetadata(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{
		ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen,
		WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow,
	}
	s := newTestService(repo)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")
	defer os.RemoveAll("./uploads")

	attachment, err := s.UploadAttachment(context.Background(), "1", testUserID, fileHeader)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attachment.TaskID != 1 {
		t.Fatalf("expected attachment task id 1, got %d", attachment.TaskID)
	}
	if attachment.FileName != "example.txt" {
		t.Fatalf("expected filename example.txt, got %q", attachment.FileName)
	}
	if _, err := os.Stat(attachment.StoragePath); err != nil {
		t.Fatalf("expected saved file at %q, got error %v", attachment.StoragePath, err)
	}
	if len(repo.attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(repo.attachments))
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

func (f *fakeRepo) GetAttachments(_ context.Context, taskID int) ([]domain.Attachment, error) {
	result := make([]domain.Attachment, 0)

	for _, attachment := range f.attachments {
		if attachment.TaskID == taskID {
			result = append(result, *attachment)
		}
	}

	return result, nil
}
