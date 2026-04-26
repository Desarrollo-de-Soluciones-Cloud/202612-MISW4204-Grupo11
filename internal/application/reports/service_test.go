package reports

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

// --- Fakes ---

type fakeReportRepo struct {
	reports map[int64]*domain.Report
	nextID  int64
}

func newFakeReportRepo() *fakeReportRepo {
	return &fakeReportRepo{reports: map[int64]*domain.Report{}, nextID: 1}
}

func (f *fakeReportRepo) Create(_ context.Context, r *domain.Report) error {
	r.ID = f.nextID
	r.CreatedAt = time.Now()
	f.nextID++
	copied := *r
	f.reports[r.ID] = &copied
	return nil
}

func (f *fakeReportRepo) FindByID(_ context.Context, id int64) (*domain.Report, error) {
	r, ok := f.reports[id]
	if !ok {
		return nil, domain.ErrReporteNoEncontrado
	}
	copied := *r
	return &copied, nil
}

func (f *fakeReportRepo) FindByProfessorAndWeek(_ context.Context, profID int64, ws time.Time) ([]domain.Report, error) {
	var out []domain.Report
	for _, r := range f.reports {
		if r.ProfessorID == profID && r.WeekStart.Equal(ws) {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeReportRepo) FindByProfessor(_ context.Context, profID int64) ([]domain.Report, error) {
	var out []domain.Report
	for _, r := range f.reports {
		if r.ProfessorID == profID {
			out = append(out, *r)
		}
	}
	return out, nil
}

type fakeAssignmentRepo struct {
	byProfessor map[int64][]domain.AssignmentWithUser
}

func (f *fakeAssignmentRepo) Create(_ context.Context, _ *domain.Assignment) error { return nil }
func (f *fakeAssignmentRepo) FindByID(_ context.Context, _ int64) (*domain.Assignment, error) {
	return nil, nil
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
func (f *fakeAssignmentRepo) FindByProfessorWithUser(_ context.Context, profID int64) ([]domain.AssignmentWithUser, error) {
	return f.byProfessor[profID], nil
}
func (f *fakeAssignmentRepo) ListAll(_ context.Context) ([]domain.Assignment, error) { return nil, nil }
func (f *fakeAssignmentRepo) Update(_ context.Context, _ *domain.Assignment) error   { return nil }

type fakeTaskRepo struct {
	byAssignmentWeek map[string][]domain.Task
}

func (f *fakeTaskRepo) Create(_ *domain.Task) error                      { return nil }
func (f *fakeTaskRepo) ListAll(_ context.Context) ([]domain.Task, error) { return nil, nil }
func (f *fakeTaskRepo) ListByUser(_ context.Context, _ int64) ([]domain.Task, error) {
	return nil, nil
}
func (f *fakeTaskRepo) ListByProfessorID(_ context.Context, _ int64) ([]domain.Task, error) {
	return nil, nil
}
func (f *fakeTaskRepo) GetByID(_ string) (*domain.Task, error) { return nil, nil }
func (f *fakeTaskRepo) GetByIDForUser(_ context.Context, _ string, _ int64) (*domain.Task, error) {
	return nil, nil
}
func (f *fakeTaskRepo) Update(_ *domain.Task) error               { return nil }
func (f *fakeTaskRepo) Delete(_ string) error                     { return nil }
func (f *fakeTaskRepo) SaveAttachment(_ *domain.Attachment) error { return nil }
func (f *fakeTaskRepo) UpdateStatus(_ *domain.Task) error         { return nil }

func (f *fakeTaskRepo) ListByAssignmentAndWeek(_ context.Context, assignmentID int64, weekStart time.Time) ([]domain.Task, error) {
	key := taskKey(assignmentID, weekStart)
	return f.byAssignmentWeek[key], nil
}

func (f *fakeTaskRepo) ListByAssignment(_ context.Context, assignmentID int64) ([]domain.Task, error) {
	var result []domain.Task
	for _, tasks := range f.byAssignmentWeek {
		for _, task := range tasks {
			if task.AssignmentId == int(assignmentID) {
				result = append(result, task)
			}
		}
	}
	return result, nil
}

func taskKey(assignmentID int64, weekStart time.Time) string {
	return time.Time.Format(weekStart, "2006-01-02") + "/" + string(rune(assignmentID+'0'))
}

type fakeAI struct {
	response string
	err      error
}

func (f *fakeAI) Summarize(_ context.Context, _ string) (string, error) {
	return f.response, f.err
}

type fakePDF struct {
	lastData *ports.PDFReportData
}

func (f *fakePDF) Generate(data ports.PDFReportData) (string, error) {
	f.lastData = &data
	return "./uploads/reports/fake.pdf", nil
}

// --- Helpers ---

var monday = time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
var tuesday = time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)

const profID int64 = 10

func sampleAssignments() map[int64][]domain.AssignmentWithUser {
	return map[int64][]domain.AssignmentWithUser{
		profID: {
			{
				Assignment: domain.Assignment{ID: 1, UserID: 3, ProfessorID: profID, RoleInAssignment: "monitor", ContractedHoursPerWeek: 8},
				UserName:   "Juan Perez",
				UserEmail:  "juan@test.com",
			},
		},
	}
}

func sampleTasks() map[string][]domain.Task {
	return map[string][]domain.Task{
		taskKey(1, monday): {
			{ID: 1, Title: "Task 1", Description: "Desc 1", Status: domain.StatusOpen, WeekStart: monday, TimeInvested: 3, AssignmentId: 1},
			{ID: 2, Title: "Task 2", Description: "Desc 2", Status: domain.StatusFinalized, WeekStart: monday, TimeInvested: 5, AssignmentId: 1},
		},
	}
}

func newTestService(assignData map[int64][]domain.AssignmentWithUser, taskData map[string][]domain.Task, ai *fakeAI) (*ReportService, *fakeReportRepo, *fakePDF) {
	reportRepo := newFakeReportRepo()
	pdfGen := &fakePDF{}
	svc := NewReportService(
		reportRepo,
		&fakeAssignmentRepo{byProfessor: assignData},
		&fakeTaskRepo{byAssignmentWeek: taskData},
		ai,
		pdfGen,
	)
	return svc, reportRepo, pdfGen
}

// --- Tests ---

func TestGenerateWeekly_WithTasks_GeneratesReport(t *testing.T) {
	svc, reportRepo, pdfGen := newTestService(sampleAssignments(), sampleTasks(), &fakeAI{response: "Great work this week."})

	reports, err := svc.GenerateWeeklyReports(context.Background(), profID, monday)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].UserName != "Juan Perez" {
		t.Fatalf("expected Juan Perez, got %q", reports[0].UserName)
	}
	if reports[0].AISummary != "Great work this week." {
		t.Fatalf("expected AI summary, got %q", reports[0].AISummary)
	}
	if len(reportRepo.reports) != 1 {
		t.Fatal("expected report stored in repo")
	}
	if pdfGen.lastData == nil {
		t.Fatal("expected PDF generator to be called")
	}
	if pdfGen.lastData.TotalHoursWorked != 8 {
		t.Fatalf("expected 8 total hours, got %d", pdfGen.lastData.TotalHoursWorked)
	}
}

func TestGenerateWeekly_NoAssignments_EmptyResult(t *testing.T) {
	svc, _, _ := newTestService(map[int64][]domain.AssignmentWithUser{}, map[string][]domain.Task{}, &fakeAI{response: ""})

	reports, err := svc.GenerateWeeklyReports(context.Background(), profID, monday)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports, got %d", len(reports))
	}
}

func TestGenerateWeekly_NoTasks_SkipsAssignment(t *testing.T) {
	svc, _, _ := newTestService(sampleAssignments(), map[string][]domain.Task{}, &fakeAI{response: ""})

	reports, err := svc.GenerateWeeklyReports(context.Background(), profID, monday)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports (no tasks), got %d", len(reports))
	}
}

func TestGenerateWeekly_AIFailure_FallbackSummary(t *testing.T) {
	svc, _, _ := newTestService(sampleAssignments(), sampleTasks(), &fakeAI{err: errors.New("ollama down")})

	reports, err := svc.GenerateWeeklyReports(context.Background(), profID, monday)
	if err != nil {
		t.Fatalf("expected success despite AI failure, got %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].AISummary != "Resumen no disponible (error en el servicio de IA)." {
		t.Fatalf("expected fallback summary, got %q", reports[0].AISummary)
	}
}

func TestGenerateWeekly_InvalidWeekStart_Rejected(t *testing.T) {
	svc, _, _ := newTestService(sampleAssignments(), sampleTasks(), &fakeAI{response: ""})

	_, err := svc.GenerateWeeklyReports(context.Background(), profID, tuesday)
	if !errors.Is(err, domain.ErrSemanaInicioNoEsLunes) {
		t.Fatalf("expected ErrSemanaInicioNoEsLunes, got %v", err)
	}
}

func TestGetReportFile_NotFound(t *testing.T) {
	svc, _, _ := newTestService(nil, nil, &fakeAI{})

	_, err := svc.GetReportFile(context.Background(), 999, profID)
	if !errors.Is(err, domain.ErrReporteNoEncontrado) {
		t.Fatalf("expected ErrReporteNoEncontrado, got %v", err)
	}
}

func TestGetReportFile_WrongProfessor_Forbidden(t *testing.T) {
	svc, reportRepo, _ := newTestService(nil, nil, &fakeAI{})
	reportRepo.reports[1] = &domain.Report{ID: 1, ProfessorID: 99, FilePath: "/some/path.pdf"}

	_, err := svc.GetReportFile(context.Background(), 1, profID)
	if !errors.Is(err, domain.ErrReporteNoAutorizado) {
		t.Fatalf("expected ErrReporteNoAutorizado, got %v", err)
	}
}

func (f *fakeTaskRepo) GetAttachments(_ context.Context, _ int) ([]domain.Attachment, error) {
	return []domain.Attachment{}, nil
}
