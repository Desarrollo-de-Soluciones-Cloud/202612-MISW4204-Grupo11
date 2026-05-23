package tasks

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type fakeStorage struct {
	saveErr error
	path    string
}

func (f *fakeStorage) Save(_ context.Context, objectPath, contentType string, data io.Reader) (ports.StoredObject, error) {
	if f.saveErr != nil {
		return ports.StoredObject{}, f.saveErr
	}
	_, _ = io.ReadAll(data)
	if f.path == "" {
		f.path = "local://" + objectPath
	}
	return ports.StoredObject{Path: f.path, ContentType: contentType}, nil
}

func (f *fakeStorage) Open(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return io.NopCloser(strings.NewReader("")), "application/octet-stream", nil
}

type fakeRepo struct {
	tasks                map[int]*domain.Task
	attachments          []*domain.Attachment
	nextID               int
	assignmentUsers      map[int]int64
	assignmentProfessors map[int]int64
	errListByUser        error
	errListByProfessor   error
	errListAll           error
	errGetByIDForUser    error
	errGetByID           error
	errGetAttachments    error
	errListByAssignment  error
	errSaveAttachment    error
	errUpdate            error
	errUpdateStatus      error
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
	if repo.errListAll != nil {
		return nil, repo.errListAll
	}
	list := make([]domain.Task, 0, len(repo.tasks))
	for _, task := range repo.tasks {
		list = append(list, *task)
	}
	return list, nil
}

func (repo *fakeRepo) ListByUser(_ context.Context, userID int64) ([]domain.Task, error) {
	if repo.errListByUser != nil {
		return nil, repo.errListByUser
	}
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
	if repo.errListByProfessor != nil {
		return nil, repo.errListByProfessor
	}
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
	if repo.errGetByID != nil {
		return nil, repo.errGetByID
	}
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
	if repo.errGetByIDForUser != nil {
		return nil, repo.errGetByIDForUser
	}
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
	if repo.errUpdate != nil {
		return repo.errUpdate
	}
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
	if repo.errSaveAttachment != nil {
		return repo.errSaveAttachment
	}
	attachment.ID = repo.nextID
	repo.nextID++
	repo.attachments = append(repo.attachments, attachment)
	return nil
}

func (repo *fakeRepo) UpdateStatus(task *domain.Task) error {
	if repo.errUpdateStatus != nil {
		return repo.errUpdateStatus
	}
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
	if repo.errListByAssignment != nil {
		return nil, repo.errListByAssignment
	}
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

func TestUpdate_InvalidTaskID(t *testing.T) {
	s := newTestService(newFakeRepo())
	err := s.Update(context.Background(), &domain.Task{ID: 0}, testUserID)
	if err == nil || !strings.Contains(err.Error(), ErrInvalidTaskID) {
		t.Fatalf("expected invalid task id error, got %v", err)
	}
}

func TestUpdate_ValidationErrors(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "old", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	tests := []struct {
		name      string
		task      domain.Task
		errSubstr string
	}{
		{"empty title", domain.Task{ID: 1, Title: "", Description: "d", Status: domain.StatusOpen, TimeInvested: 1, AssignmentId: 1}, ErrTitleRequired},
		{"empty desc", domain.Task{ID: 1, Title: "t", Description: "", Status: domain.StatusOpen, TimeInvested: 1, AssignmentId: 1}, ErrDescriptionRequired},
		{"empty status", domain.Task{ID: 1, Title: "t", Description: "d", Status: "", TimeInvested: 1, AssignmentId: 1}, ErrStatusRequired},
		{"invalid time", domain.Task{ID: 1, Title: "t", Description: "d", Status: domain.StatusOpen, TimeInvested: 0, AssignmentId: 1}, ErrTimeInvestedMustBeGreater},
		{"too many hours", domain.Task{ID: 1, Title: "t", Description: "d", Status: domain.StatusOpen, TimeInvested: 23, AssignmentId: 1}, ErrMaxTimeInvestedPerTask},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Update(context.Background(), &tc.task, testUserID)
			if err == nil || !strings.Contains(err.Error(), tc.errSubstr) {
				t.Fatalf("expected %q error, got %v", tc.errSubstr, err)
			}
		})
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

func TestPartialUpdate_ValidationErrors(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	empty := "   "
	_, err := s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Title: &empty})
	if err == nil || !strings.Contains(err.Error(), ErrTitleRequired) {
		t.Fatalf("expected title required, got %v", err)
	}

	repo.tasks[1].Title = "task"
	zero := 0
	_, err = s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{TimeInvested: &zero})
	if err == nil || !strings.Contains(err.Error(), ErrTimeInvestedMustBeGreater) {
		t.Fatalf("expected time invested validation, got %v", err)
	}

	repo.tasks[1].TimeInvested = 3
	tooMany := 23
	_, err = s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{TimeInvested: &tooMany})
	if err == nil || !strings.Contains(err.Error(), ErrMaxTimeInvestedPerTask) {
		t.Fatalf("expected max hours validation, got %v", err)
	}

	repo.tasks[1].TimeInvested = 3
	emptyDesc := "  "
	_, err = s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Description: &emptyDesc})
	if err == nil || !strings.Contains(err.Error(), ErrDescriptionRequired) {
		t.Fatalf("expected description required, got %v", err)
	}

	repo.tasks[1].Description = "desc"
	emptyStatus := domain.Status(" ")
	_, err = s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Status: &emptyStatus})
	if err == nil || !strings.Contains(err.Error(), ErrStatusRequired) {
		t.Fatalf("expected status required, got %v", err)
	}
}

func TestPartialUpdate_RepositoryUpdateError(t *testing.T) {
	repo := newFakeRepo()
	repo.errUpdate = errors.New("update failed")
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 3, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	obs := "new observations"
	_, err := s.PartialUpdate(context.Background(), "1", testUserID, UpdateTaskInput{Observations: &obs})
	if err == nil || !strings.Contains(err.Error(), "update failed") {
		t.Fatalf("expected repo update error, got %v", err)
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

func TestUpdateStatus_InvalidInput(t *testing.T) {
	s := newTestService(newFakeRepo())
	err := s.UpdateStatus(context.Background(), &domain.Task{ID: 0}, testUserID)
	if err == nil || !strings.Contains(err.Error(), ErrInvalidTaskID) {
		t.Fatalf("expected invalid id error, got %v", err)
	}
}

func TestUpdateStatus_EmptyStatusRejected(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	err := s.UpdateStatus(context.Background(), &domain.Task{ID: 1, Status: ""}, testUserID)
	if err == nil || !strings.Contains(err.Error(), ErrStatusRequired) {
		t.Fatalf("expected status required error, got %v", err)
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

func TestUploadAttachment_InvalidTaskID(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")

	_, err := s.UploadAttachment(context.Background(), "abc", testUserID, fileHeader)
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Fatalf("expected task not found error for malformed id, got %v", err)
	}
}

func TestUploadAttachment_SaveAttachmentError(t *testing.T) {
	repo := newFakeRepo()
	repo.errSaveAttachment = errors.New("save attachment failed")
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")
	defer os.RemoveAll("./uploads")

	_, err := s.UploadAttachment(context.Background(), "1", testUserID, fileHeader)
	if err == nil || !strings.Contains(err.Error(), "save attachment failed") {
		t.Fatalf("expected save attachment error, got %v", err)
	}
}

func TestUploadAttachment_FileOpenError(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	_, err := s.UploadAttachment(context.Background(), "1", testUserID, &multipart.FileHeader{})
	if err == nil || !strings.Contains(err.Error(), "could not open file") {
		t.Fatalf("expected open file error, got %v", err)
	}
}

func TestUploadAttachment_WithStorage_SaveSuccessAndFailure(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	storage := &fakeStorage{path: "local://attachments/custom.txt"}
	s.WithFileStorage(storage)
	fileHeader := createMultipartFileHeader(t, "file", "example.txt", "hello world")

	attachment, err := s.UploadAttachment(context.Background(), "1", testUserID, fileHeader)
	if err != nil {
		t.Fatalf("expected success with storage, got %v", err)
	}
	if attachment.StoragePath != "local://attachments/custom.txt" {
		t.Fatalf("expected storage path from storage backend, got %q", attachment.StoragePath)
	}

	storage.saveErr = errors.New("storage save failed")
	_, err = s.UploadAttachment(context.Background(), "1", testUserID, fileHeader)
	if err == nil || !strings.Contains(err.Error(), "could not save file in storage") {
		t.Fatalf("expected wrapped storage error, got %v", err)
	}
}

func TestGetAttachments_OkAndErrorPath(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	repo.attachments = []*domain.Attachment{{TaskID: 1, FileName: "a.txt"}}
	s := newTestService(repo)

	items, err := s.GetAttachments(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(items))
	}

	repo.errGetByID = errors.New("task lookup failed")
	_, err = s.GetAttachments(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "task lookup failed") {
		t.Fatalf("expected task lookup error, got %v", err)
	}

	repo.errGetByID = nil
	repo.errGetAttachments = errors.New("attachment lookup failed")
	_, err = s.GetAttachments(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "attachment lookup failed") {
		t.Fatalf("expected attachment lookup error, got %v", err)
	}
}

func TestListMethods_DelegateToRepository(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	s := newTestService(repo)

	if _, err := s.ListForUser(context.Background(), testUserID); err != nil {
		t.Fatalf("expected list for user success, got %v", err)
	}
	if _, err := s.ListForProfessor(context.Background(), 10); err != nil {
		t.Fatalf("expected list for professor success, got %v", err)
	}
	if _, err := s.ListAllForAdmin(context.Background()); err != nil {
		t.Fatalf("expected list all success, got %v", err)
	}
	if _, err := s.GetByIDForUser(context.Background(), "1", testUserID); err != nil {
		t.Fatalf("expected get by id for user success, got %v", err)
	}

	repo.errListByUser = errors.New("list by user failed")
	if _, err := s.ListForUser(context.Background(), testUserID); err == nil {
		t.Fatal("expected list for user error")
	}
	repo.errListByProfessor = errors.New("list by professor failed")
	if _, err := s.ListForProfessor(context.Background(), 10); err == nil {
		t.Fatal("expected list for professor error")
	}
	repo.errListAll = errors.New("list all failed")
	if _, err := s.ListAllForAdmin(context.Background()); err == nil {
		t.Fatal("expected list all error")
	}
	repo.errGetByIDForUser = errors.New("get by id for user failed")
	if _, err := s.GetByIDForUser(context.Background(), "1", testUserID); err == nil {
		t.Fatal("expected get by id for user error")
	}
}

func TestListByAssignment_AccessControlAndErrors(t *testing.T) {
	repo := newFakeRepo()
	repo.tasks[1] = &domain.Task{ID: 1, Title: "task", Description: "desc", Status: domain.StatusOpen, WeekStart: thisMonday, TimeInvested: 2, AssignmentId: 1, TimeRegistered: fixedNow}
	assign := newFakeAssignmentRepo()
	s := NewTaskService(repo, assign)

	if _, err := s.ListByAssignment(context.Background(), 1, 1); err != nil {
		t.Fatalf("expected owner access, got %v", err)
	}
	if _, err := s.ListByAssignment(context.Background(), 1, 10); err != nil {
		t.Fatalf("expected professor access, got %v", err)
	}
	if _, err := s.ListByAssignment(context.Background(), 1, 999); !errors.Is(err, domain.ErrAssignmentNotOwned) {
		t.Fatalf("expected ErrAssignmentNotOwned, got %v", err)
	}

	assign.byID = map[int64]*domain.Assignment{}
	if _, err := s.ListByAssignment(context.Background(), 1, 1); !errors.Is(err, domain.ErrVinculacionNoEncontrada) {
		t.Fatalf("expected assignment lookup error, got %v", err)
	}

	assign.byID[1] = &domain.Assignment{ID: 1, UserID: 1, ProfessorID: 10}
	repo.errListByAssignment = errors.New("list by assignment failed")
	if _, err := s.ListByAssignment(context.Background(), 1, 1); err == nil || !strings.Contains(err.Error(), "list by assignment failed") {
		t.Fatalf("expected list by assignment error, got %v", err)
	}
}

func TestWithFileStorage_SetsStorage(t *testing.T) {
	repo := newFakeRepo()
	s := newTestService(repo)

	stored := s.WithFileStorage(nil)
	if stored != s {
		t.Fatal("expected fluent API to return same service pointer")
	}
}

func TestSaveFile_WritesMultipartToDestination(t *testing.T) {
	fileHeader := createMultipartFileHeader(t, "file", "sample.txt", "content")
	dir := t.TempDir()
	dst := dir + "/uploads/sample.txt"

	if err := saveFile(fileHeader, dst); err != nil {
		t.Fatalf("expected saveFile success, got %v", err)
	}
	bytes, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("expected written file, got %v", err)
	}
	if string(bytes) != "content" {
		t.Fatalf("unexpected file content %q", string(bytes))
	}
}

func TestSaveFile_OpenError(t *testing.T) {
	err := saveFile(&multipart.FileHeader{}, "./tmp/any.txt")
	if err == nil {
		t.Fatal("expected saveFile to fail when file header cannot be opened")
	}
}

func TestSaveFileFallback_InvalidDestination(t *testing.T) {
	fileHeader := createMultipartFileHeader(t, "file", "sample.txt", "content")
	reader, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer reader.Close()

	err = saveFileFallback(reader, "\x00invalid/path.txt")
	if err == nil {
		t.Fatal("expected saveFileFallback to fail with invalid destination")
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
	if f.errGetAttachments != nil {
		return nil, f.errGetAttachments
	}
	result := make([]domain.Attachment, 0)

	for _, attachment := range f.attachments {
		if attachment.TaskID == taskID {
			result = append(result, *attachment)
		}
	}

	return result, nil
}
