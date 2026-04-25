package handlers

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

func newJSONContext(method, path, body string, authUID any) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req
	if authUID != nil {
		c.Set("authUserID", authUID)
	}
	return c, w
}

type memoryPeriodRepo struct {
	mu      sync.Mutex
	byID    map[int64]*domain.AcademicPeriod
	nextID  int64
	listErr error
}

func newMemoryPeriodRepo() *memoryPeriodRepo {
	return &memoryPeriodRepo{byID: make(map[int64]*domain.AcademicPeriod), nextID: 1}
}

func (m *memoryPeriodRepo) FindByID(_ context.Context, id int64) (*domain.AcademicPeriod, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrPeriodoNoEncontrado
	}
	cp := *p
	return &cp, nil
}

func (m *memoryPeriodRepo) Create(_ context.Context, p *domain.AcademicPeriod) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.ID = m.nextID
	m.nextID++
	cp := *p
	m.byID[p.ID] = &cp
	return nil
}

func (m *memoryPeriodRepo) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.AcademicPeriod, 0, len(m.byID))
	for _, p := range m.byID {
		out = append(out, *p)
	}
	return out, nil
}

func (m *memoryPeriodRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.byID[id]
	if !ok {
		return domain.ErrPeriodoNoEncontrado
	}
	p.Status = status
	return nil
}

type memorySpaceRepo struct {
	mu            sync.Mutex
	byID          map[int64]*domain.AcademicSpace
	nextID        int64
	listErr       error
	findByProfErr error
}

func newMemorySpaceRepo() *memorySpaceRepo {
	return &memorySpaceRepo{byID: make(map[int64]*domain.AcademicSpace), nextID: 1}
}

func (m *memorySpaceRepo) Create(_ context.Context, s *domain.AcademicSpace) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.ID = m.nextID
	m.nextID++
	cp := *s
	m.byID[s.ID] = &cp
	return nil
}

func (m *memorySpaceRepo) FindByID(_ context.Context, id int64) (*domain.AcademicSpace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrEspacioNoEncontrado
	}
	cp := *s
	return &cp, nil
}

func (m *memorySpaceRepo) FindByProfessor(_ context.Context, professorID int64) ([]domain.AcademicSpace, error) {
	if m.findByProfErr != nil {
		return nil, m.findByProfErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.AcademicSpace
	for _, s := range m.byID {
		if s.ProfessorID == professorID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (m *memorySpaceRepo) ListAll(_ context.Context) ([]domain.AcademicSpace, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.AcademicSpace, 0, len(m.byID))
	for _, s := range m.byID {
		out = append(out, *s)
	}
	return out, nil
}

func (m *memorySpaceRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.byID[id]
	if !ok {
		return domain.ErrEspacioNoEncontrado
	}
	s.Status = status
	return nil
}

type memoryAssignmentRepo struct {
	mu              sync.Mutex
	byID            map[int64]*domain.Assignment
	nextID          int64
	findByUserErr   error
	listAllErr      error
	findByProfWUErr error
}

func newMemoryAssignmentRepo() *memoryAssignmentRepo {
	return &memoryAssignmentRepo{byID: make(map[int64]*domain.Assignment), nextID: 1}
}

func (m *memoryAssignmentRepo) Create(_ context.Context, a *domain.Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a.ID = m.nextID
	m.nextID++
	cp := *a
	m.byID[a.ID] = &cp
	return nil
}

func (m *memoryAssignmentRepo) FindByID(_ context.Context, id int64) (*domain.Assignment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrVinculacionNoEncontrada
	}
	cp := *a
	return &cp, nil
}

func (m *memoryAssignmentRepo) FindBySpace(_ context.Context, spaceID int64) ([]domain.Assignment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Assignment
	for _, a := range m.byID {
		if a.AcademicSpaceID == spaceID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (m *memoryAssignmentRepo) FindByUser(_ context.Context, userID int64) ([]domain.Assignment, error) {
	if m.findByUserErr != nil {
		return nil, m.findByUserErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Assignment
	for _, a := range m.byID {
		if a.UserID == userID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (m *memoryAssignmentRepo) ExistsByUserSpaceRole(_ context.Context, userID, spaceID int64, role string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.byID {
		if a.UserID == userID && a.AcademicSpaceID == spaceID && a.RoleInAssignment == role {
			return true, nil
		}
	}
	return false, nil
}

func (m *memoryAssignmentRepo) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}

func (m *memoryAssignmentRepo) FindByProfessorWithUser(_ context.Context, professorID int64) ([]domain.AssignmentWithUser, error) {
	if m.findByProfWUErr != nil {
		return nil, m.findByProfWUErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.AssignmentWithUser
	for _, a := range m.byID {
		if a.ProfessorID == professorID {
			out = append(out, domain.AssignmentWithUser{
				Assignment: *a,
				UserName:   "Test User",
				UserEmail:  "u@test.com",
			})
		}
	}
	return out, nil
}

func (m *memoryAssignmentRepo) ListAll(_ context.Context) ([]domain.Assignment, error) {
	if m.listAllErr != nil {
		return nil, m.listAllErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Assignment, 0, len(m.byID))
	for _, a := range m.byID {
		out = append(out, *a)
	}
	return out, nil
}

func (m *memoryAssignmentRepo) Update(_ context.Context, a *domain.Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[a.ID]; !ok {
		return domain.ErrVinculacionNoEncontrada
	}
	cp := *a
	m.byID[a.ID] = &cp
	return nil
}

type memoryReportRepo struct {
	mu                        sync.Mutex
	byID                      map[int64]*domain.Report
	next                      int64
	createErr                 error
	findByProfessorErr        error
	findByProfessorAndWeekErr error
}

func newMemoryReportRepo() *memoryReportRepo {
	return &memoryReportRepo{byID: make(map[int64]*domain.Report), next: 1}
}

func (m *memoryReportRepo) Create(_ context.Context, r *domain.Report) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	r.ID = m.next
	m.next++
	cp := *r
	m.byID[r.ID] = &cp
	return nil
}

func (m *memoryReportRepo) FindByID(_ context.Context, id int64) (*domain.Report, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrReporteNoEncontrado
	}
	cp := *r
	return &cp, nil
}

func (m *memoryReportRepo) FindByProfessorAndWeek(_ context.Context, professorID int64, weekStart time.Time) ([]domain.Report, error) {
	if m.findByProfessorAndWeekErr != nil {
		return nil, m.findByProfessorAndWeekErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Report
	ws := weekStart.UTC().Truncate(24 * time.Hour)
	for _, r := range m.byID {
		rws := r.WeekStart.UTC().Truncate(24 * time.Hour)
		if r.ProfessorID == professorID && rws.Equal(ws) {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *memoryReportRepo) FindByProfessor(_ context.Context, professorID int64) ([]domain.Report, error) {
	if m.findByProfessorErr != nil {
		return nil, m.findByProfessorErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.Report
	for _, r := range m.byID {
		if r.ProfessorID == professorID {
			out = append(out, *r)
		}
	}
	return out, nil
}

type handlerTaskRepo struct {
	tasks                map[int]*domain.Task
	attachments          []*domain.Attachment
	nextID               int
	assignmentUsers      map[int]int64
	assignmentProfessors map[int]int64
	listAllErr           error
	listByUserErr        error
	listByProfErr        error
	getByUserAfterOK     error
	saveAttachmentErr    error
}

func newHandlerTaskRepo() *handlerTaskRepo {
	return &handlerTaskRepo{
		tasks:                map[int]*domain.Task{},
		attachments:          []*domain.Attachment{},
		nextID:               1,
		assignmentUsers:      map[int]int64{1: 1},
		assignmentProfessors: map[int]int64{1: 10},
	}
}

func (repo *handlerTaskRepo) Create(task *domain.Task) error {
	task.ID = repo.nextID
	repo.nextID++
	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (repo *handlerTaskRepo) ListAll(_ context.Context) ([]domain.Task, error) {
	if repo.listAllErr != nil {
		return nil, repo.listAllErr
	}
	list := make([]domain.Task, 0, len(repo.tasks))
	for _, task := range repo.tasks {
		list = append(list, *task)
	}
	return list, nil
}

func (repo *handlerTaskRepo) ListByUser(_ context.Context, userID int64) ([]domain.Task, error) {
	if repo.listByUserErr != nil {
		return nil, repo.listByUserErr
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

func (repo *handlerTaskRepo) ListByProfessorID(_ context.Context, professorID int64) ([]domain.Task, error) {
	if repo.listByProfErr != nil {
		return nil, repo.listByProfErr
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

func (repo *handlerTaskRepo) GetByID(id string) (*domain.Task, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errHandlerTaskLegacyNotFound
	}
	task, ok := repo.tasks[intID]
	if !ok {
		return nil, errHandlerTaskLegacyNotFound
	}
	copied := *task
	return &copied, nil
}

func (repo *handlerTaskRepo) GetByIDForUser(_ context.Context, id string, userID int64) (*domain.Task, error) {
	task, err := repo.GetByID(id)
	if err != nil {
		return nil, domain.ErrTaskNotFound
	}
	owner, ok := repo.assignmentUsers[task.AssignmentId]
	if !ok || owner != userID {
		return nil, domain.ErrTaskNotFound
	}
	if repo.getByUserAfterOK != nil {
		return nil, repo.getByUserAfterOK
	}
	return task, nil
}

func (repo *handlerTaskRepo) Update(task *domain.Task) error {
	if _, ok := repo.tasks[task.ID]; !ok {
		return errHandlerTaskLegacyNotFound
	}
	copied := *task
	repo.tasks[task.ID] = &copied
	return nil
}

func (repo *handlerTaskRepo) Delete(id string) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return errHandlerTaskLegacyNotFound
	}
	if _, ok := repo.tasks[intID]; !ok {
		return errHandlerTaskLegacyNotFound
	}
	delete(repo.tasks, intID)
	return nil
}

func (repo *handlerTaskRepo) SaveAttachment(attachment *domain.Attachment) error {
	if repo.saveAttachmentErr != nil {
		return repo.saveAttachmentErr
	}
	attachment.ID = repo.nextID
	repo.nextID++
	repo.attachments = append(repo.attachments, attachment)
	return nil
}

func (repo *handlerTaskRepo) UpdateStatus(task *domain.Task) error {
	stored, ok := repo.tasks[task.ID]
	if !ok {
		return errHandlerTaskLegacyNotFound
	}
	stored.Status = task.Status
	return nil
}

func (repo *handlerTaskRepo) ListByAssignmentAndWeek(_ context.Context, _ int64, _ time.Time) ([]domain.Task, error) {
	return nil, nil
}

var errHandlerTaskLegacyNotFound = errors.New("task not found")

type taskAssignmentLookup struct {
	byID map[int64]*domain.Assignment
}

func (t *taskAssignmentLookup) Create(_ context.Context, a *domain.Assignment) error {
	t.byID[a.ID] = a
	return nil
}
func (t *taskAssignmentLookup) FindByID(_ context.Context, id int64) (*domain.Assignment, error) {
	a, ok := t.byID[id]
	if !ok {
		return nil, domain.ErrVinculacionNoEncontrada
	}
	cp := *a
	return &cp, nil
}
func (t *taskAssignmentLookup) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (t *taskAssignmentLookup) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (t *taskAssignmentLookup) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return false, nil
}
func (t *taskAssignmentLookup) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}
func (t *taskAssignmentLookup) FindByProfessorWithUser(_ context.Context, _ int64) ([]domain.AssignmentWithUser, error) {
	return nil, nil
}
func (t *taskAssignmentLookup) ListAll(_ context.Context) ([]domain.Assignment, error) {
	return nil, nil
}
func (t *taskAssignmentLookup) Update(_ context.Context, _ *domain.Assignment) error { return nil }

func reportTaskKey(assignmentID int64, weekStart time.Time) string {
	return weekStart.Format("2006-01-02") + "/" + strconv.FormatInt(assignmentID, 10)
}

type reportFakeAssignmentRepo struct {
	byProfessor map[int64][]domain.AssignmentWithUser
	err         error
}

func (f *reportFakeAssignmentRepo) Create(_ context.Context, _ *domain.Assignment) error { return nil }
func (f *reportFakeAssignmentRepo) FindByID(_ context.Context, _ int64) (*domain.Assignment, error) {
	return nil, nil
}
func (f *reportFakeAssignmentRepo) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (f *reportFakeAssignmentRepo) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (f *reportFakeAssignmentRepo) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return false, nil
}
func (f *reportFakeAssignmentRepo) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}
func (f *reportFakeAssignmentRepo) FindByProfessorWithUser(_ context.Context, profID int64) ([]domain.AssignmentWithUser, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byProfessor[profID], nil
}
func (f *reportFakeAssignmentRepo) ListAll(_ context.Context) ([]domain.Assignment, error) {
	return nil, nil
}
func (f *reportFakeAssignmentRepo) Update(_ context.Context, _ *domain.Assignment) error { return nil }

type reportFakeTaskRepo struct {
	byAssignmentWeek map[string][]domain.Task
	err              error
}

func (f *reportFakeTaskRepo) Create(_ *domain.Task) error                      { return nil }
func (f *reportFakeTaskRepo) ListAll(_ context.Context) ([]domain.Task, error) { return nil, nil }
func (f *reportFakeTaskRepo) ListByUser(_ context.Context, _ int64) ([]domain.Task, error) {
	return nil, nil
}
func (f *reportFakeTaskRepo) ListByProfessorID(_ context.Context, _ int64) ([]domain.Task, error) {
	return nil, nil
}
func (f *reportFakeTaskRepo) GetByID(_ string) (*domain.Task, error) { return nil, nil }
func (f *reportFakeTaskRepo) GetByIDForUser(_ context.Context, _ string, _ int64) (*domain.Task, error) {
	return nil, nil
}
func (f *reportFakeTaskRepo) Update(_ *domain.Task) error               { return nil }
func (f *reportFakeTaskRepo) Delete(_ string) error                     { return nil }
func (f *reportFakeTaskRepo) SaveAttachment(_ *domain.Attachment) error { return nil }
func (f *reportFakeTaskRepo) UpdateStatus(_ *domain.Task) error         { return nil }
func (f *reportFakeTaskRepo) ListByAssignmentAndWeek(_ context.Context, assignmentID int64, weekStart time.Time) ([]domain.Task, error) {
	if f.err != nil {
		return nil, f.err
	}
	key := reportTaskKey(assignmentID, weekStart)
	return f.byAssignmentWeek[key], nil
}

type reportFakeAI struct {
	response string
	err      error
}

func (f *reportFakeAI) Summarize(_ context.Context, _ string) (string, error) {
	return f.response, f.err
}

type handlerReportPDFStub struct{}

func (handlerReportPDFStub) Generate(_ ports.PDFReportData) (string, error) {
	return "/tmp/handler-test-report.pdf", nil
}

type handlerReportPDFFail struct{}

func (handlerReportPDFFail) Generate(_ ports.PDFReportData) (string, error) {
	return "", errors.New("pdf generation failed")
}

type handlerUserRepo struct {
	mu          sync.Mutex
	creds       map[string]*domain.UserCredentials
	users       []domain.User
	count       int64
	nextID      int64
	listErr     error
	countErr    error
	emailExists map[string]bool
}

func newHandlerUserRepo() *handlerUserRepo {
	return &handlerUserRepo{
		creds:       make(map[string]*domain.UserCredentials),
		users:       nil,
		nextID:      1,
		emailExists: make(map[string]bool),
	}
}

func (h *handlerUserRepo) FindCredentialsByEmail(_ context.Context, email string) (*domain.UserCredentials, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.creds[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (h *handlerUserRepo) CreateUser(_ context.Context, name, email, passwordHash string, roleNames []string) (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextID
	h.nextID++
	emailKey := strings.ToLower(strings.TrimSpace(email))
	h.creds[emailKey] = &domain.UserCredentials{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Roles:        append([]string(nil), roleNames...),
	}
	h.users = append(h.users, domain.User{ID: id, Name: name, Email: email, Roles: append([]string(nil), roleNames...)})
	h.emailExists[emailKey] = true
	h.count++
	return id, nil
}

func (h *handlerUserRepo) ListUsers(_ context.Context) ([]domain.User, error) {
	if h.listErr != nil {
		return nil, h.listErr
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]domain.User, len(h.users))
	copy(out, h.users)
	return out, nil
}

func (h *handlerUserRepo) ListUsersByRole(_ context.Context, _ string) ([]domain.User, error) {
	if h.listErr != nil {
		return nil, h.listErr
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]domain.User, len(h.users))
	copy(out, h.users)
	return out, nil
}

func (h *handlerUserRepo) EmailExists(_ context.Context, email string) (bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := strings.ToLower(strings.TrimSpace(email))
	return h.emailExists[key], nil
}

func (h *handlerUserRepo) CountUsers(_ context.Context) (int64, error) {
	if h.countErr != nil {
		return 0, h.countErr
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count, nil
}

var _ ports.UserRepository = (*handlerUserRepo)(nil)

type overviewUserStub struct {
	users []domain.User
	err   error
}

func (s *overviewUserStub) FindCredentialsByEmail(_ context.Context, _ string) (*domain.UserCredentials, error) {
	return nil, nil
}
func (s *overviewUserStub) CreateUser(_ context.Context, _, _, _ string, _ []string) (int64, error) {
	return 0, nil
}
func (s *overviewUserStub) ListUsers(_ context.Context) ([]domain.User, error) {
	return s.users, s.err
}
func (s *overviewUserStub) ListUsersByRole(_ context.Context, _ string) ([]domain.User, error) {
	return s.users, s.err
}
func (s *overviewUserStub) EmailExists(_ context.Context, _ string) (bool, error) { return false, nil }
func (s *overviewUserStub) CountUsers(_ context.Context) (int64, error)           { return 0, nil }

type overviewPeriodStub struct {
	periods []domain.AcademicPeriod
	err     error
}

func (s *overviewPeriodStub) FindByID(_ context.Context, _ int64) (*domain.AcademicPeriod, error) {
	return nil, nil
}
func (s *overviewPeriodStub) Create(_ context.Context, _ *domain.AcademicPeriod) error { return nil }
func (s *overviewPeriodStub) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	return s.periods, s.err
}
func (s *overviewPeriodStub) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }

type overviewSpaceStub struct {
	spaces []domain.AcademicSpace
	err    error
}

func (s *overviewSpaceStub) Create(_ context.Context, _ *domain.AcademicSpace) error { return nil }
func (s *overviewSpaceStub) FindByID(_ context.Context, _ int64) (*domain.AcademicSpace, error) {
	return nil, nil
}
func (s *overviewSpaceStub) FindByProfessor(_ context.Context, _ int64) ([]domain.AcademicSpace, error) {
	return nil, nil
}
func (s *overviewSpaceStub) ListAll(_ context.Context) ([]domain.AcademicSpace, error) {
	return s.spaces, s.err
}
func (s *overviewSpaceStub) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }

type overviewAssignStub struct {
	assignments []domain.Assignment
	err         error
}

func (s *overviewAssignStub) Create(_ context.Context, _ *domain.Assignment) error { return nil }
func (s *overviewAssignStub) FindByID(_ context.Context, _ int64) (*domain.Assignment, error) {
	return nil, nil
}
func (s *overviewAssignStub) FindBySpace(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (s *overviewAssignStub) FindByUser(_ context.Context, _ int64) ([]domain.Assignment, error) {
	return nil, nil
}
func (s *overviewAssignStub) ExistsByUserSpaceRole(_ context.Context, _, _ int64, _ string) (bool, error) {
	return false, nil
}
func (s *overviewAssignStub) FindActiveByUserAndRole(_ context.Context, _ int64, _ string) ([]domain.Assignment, error) {
	return nil, nil
}
func (s *overviewAssignStub) FindByProfessorWithUser(_ context.Context, _ int64) ([]domain.AssignmentWithUser, error) {
	return nil, nil
}
func (s *overviewAssignStub) ListAll(_ context.Context) ([]domain.Assignment, error) {
	return s.assignments, s.err
}
func (s *overviewAssignStub) Update(_ context.Context, _ *domain.Assignment) error { return nil }

type overviewTaskStub struct {
	tasks []domain.Task
	err   error
}

func (s *overviewTaskStub) ListAll(_ context.Context) ([]domain.Task, error) {
	return s.tasks, s.err
}
