package handlers

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/admin"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/auth"
	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
	appspaces "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	apptasks "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/tasks"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/users"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/gin-gonic/gin"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

var handlerTestMonday = time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
var handlerTaskNow = time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC) // miércoles

func seedOpenPeriodAndSpace(t *testing.T) (*memoryPeriodRepo, *memorySpaceRepo, *memoryAssignmentRepo) {
	t.Helper()
	periods := newMemoryPeriodRepo()
	spaces := newMemorySpaceRepo()
	assigns := newMemoryAssignmentRepo()
	ctx := context.Background()
	p := &domain.AcademicPeriod{
		Code:      "2026-1",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:    "active",
	}
	if err := periods.Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	s := &domain.AcademicSpace{
		Name:             "Curso X",
		Type:             domain.SpaceTypeCourse,
		AcademicPeriodID: 1,
		ProfessorID:      10,
		StartDate:        time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		Status:           domain.SpaceStatusActive,
	}
	if err := spaces.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	return periods, spaces, assigns
}

func TestAuth_PostLogin_InvalidBody(t *testing.T) {
	h := &Auth{Login: &auth.LoginService{Users: newHandlerUserRepo(), Secret: []byte(testJWTSecret)}}
	c, w := newJSONContext(http.MethodPost, "/login", `{}`, nil)
	h.PostLogin(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuth_PostLogin_InvalidCredentials(t *testing.T) {
	h := &Auth{Login: &auth.LoginService{Users: newHandlerUserRepo(), Secret: []byte(testJWTSecret)}}
	c, w := newJSONContext(http.MethodPost, "/login", `{"email":"a@b.co","password":"secret1234"}`, nil)
	h.PostLogin(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAuth_PostLogin_OK(t *testing.T) {
	repo := newHandlerUserRepo()
	adminSvc := &users.AdminService{Users: repo}
	ctx := context.Background()
	if _, err := adminSvc.Create(ctx, "U", "u@test.com", "secret1234", []string{domain.RolMonitor}); err != nil {
		t.Fatal(err)
	}
	h := &Auth{Login: &auth.LoginService{Users: repo, Secret: []byte(testJWTSecret)}}
	c, w := newJSONContext(http.MethodPost, "/login", `{"email":"u@test.com","password":"secret1234"}`, nil)
	h.PostLogin(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuth_PostLogin_MisconfiguredSecret(t *testing.T) {
	repo := newHandlerUserRepo()
	adminSvc := &users.AdminService{Users: repo}
	ctx := context.Background()
	if _, err := adminSvc.Create(ctx, "U", "u@test.com", "secret1234", []string{domain.RolMonitor}); err != nil {
		t.Fatal(err)
	}
	h := &Auth{Login: &auth.LoginService{Users: repo, Secret: nil}}
	c, w := newJSONContext(http.MethodPost, "/login", `{"email":"u@test.com","password":"secret1234"}`, nil)
	h.PostLogin(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAdmin_GetOverview_OK(t *testing.T) {
	svc := admin.NewPlatformOverviewService(
		&overviewUserStub{users: []domain.User{{ID: 1, Name: "A", Email: "a@x.co", Roles: []string{"administrador"}}}},
		&overviewPeriodStub{periods: []domain.AcademicPeriod{{ID: 1, Code: "2026-1", Status: "active"}}},
		&overviewSpaceStub{spaces: []domain.AcademicSpace{{ID: 1, Name: "S", ProfessorID: 2}}},
		&overviewAssignStub{assignments: []domain.Assignment{{ID: 1, UserID: 3, AcademicSpaceID: 1}}},
		&overviewTaskStub{tasks: []domain.Task{{ID: 1, Title: "T", AssignmentId: 1}}},
	)
	h := NewAdminHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/admin/overview", "", nil)
	h.GetOverview(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAdmin_GetOverview_Error(t *testing.T) {
	want := errors.New("list fail")
	svc := admin.NewPlatformOverviewService(
		&overviewUserStub{err: want},
		&overviewPeriodStub{},
		&overviewSpaceStub{},
		&overviewAssignStub{},
		&overviewTaskStub{},
	)
	h := NewAdminHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/admin/overview", "", nil)
	h.GetOverview(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicPeriod_Create_InvalidJSON(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/periods", `{`, nil)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicPeriod_Create_InvalidStartDate(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	body := `{"code":"x","start_date":"bad","end_date":"2026-12-31"}`
	c, w := newJSONContext(http.MethodPost, "/periods", body, nil)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicPeriod_Create_UnprocessableDates(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	body := `{"code":"x","start_date":"2026-12-31","end_date":"2026-01-01"}`
	c, w := newJSONContext(http.MethodPost, "/periods", body, nil)
	h.Create(c)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAcademicPeriod_Create_OK(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	body := `{"code":"2026-1","start_date":"2026-01-01","end_date":"2026-12-31"}`
	c, w := newJSONContext(http.MethodPost, "/periods", body, nil)
	h.Create(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAcademicPeriod_List_OK(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/periods", "", nil)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicPeriod_Close_InvalidID(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/periods/x/close", "", nil)
	c.AddParam("id", "x")
	h.Close(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicPeriod_Close_NotFound(t *testing.T) {
	svc := appspaces.NewAcademicPeriodService(newMemoryPeriodRepo())
	h := NewAcademicPeriodHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/periods/99/close", "", nil)
	c.AddParam("id", "99")
	h.Close(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAcademicSpace_Create_Unauthorized(t *testing.T) {
	periods, spaces, _ := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAcademicSpaceService(spaces, periods)
	h := NewAcademicSpaceHandler(svc)
	body := `{"name":"N","type":"course","academic_period_id":1,"start_date":"2026-03-15","end_date":"2026-04-15"}`
	c, w := newJSONContext(http.MethodPost, "/spaces", body, nil)
	h.Create(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicSpace_Create_OK(t *testing.T) {
	periods, spaces, _ := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAcademicSpaceService(spaces, periods)
	h := NewAcademicSpaceHandler(svc)
	body := `{"name":"N","type":"course","academic_period_id":1,"start_date":"2026-03-15","end_date":"2026-04-15"}`
	c, w := newJSONContext(http.MethodPost, "/spaces", body, int64(10))
	h.Create(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAcademicSpace_List_OK(t *testing.T) {
	periods, spaces, _ := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAcademicSpaceService(spaces, periods)
	h := NewAcademicSpaceHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/spaces", "", int64(10))
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicSpace_Get_Forbidden(t *testing.T) {
	periods, spaces, _ := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAcademicSpaceService(spaces, periods)
	h := NewAcademicSpaceHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/spaces/1", "", int64(99))
	c.AddParam("id", "1")
	h.Get(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAcademicSpace_ListAllForAdmin_OK(t *testing.T) {
	periods, spaces, _ := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAcademicSpaceService(spaces, periods)
	h := NewAcademicSpaceHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/admin/spaces", "", nil)
	h.ListAllForAdmin(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAcademicSpace_Close_OK(t *testing.T) {
	periods, spaces, _ := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAcademicSpaceService(spaces, periods)
	h := NewAcademicSpaceHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/spaces/1/close", "", int64(10))
	c.AddParam("id", "1")
	h.Close(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAssignment_Create_Unauthorized(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/spaces/1/assignments", `{"user_id":5,"role_in_assignment":"monitor","contracted_hours_per_week":8}`, nil)
	c.AddParam("id", "1")
	h.Create(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAssignment_Create_OK(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	body := `{"user_id":5,"role_in_assignment":"monitor","contracted_hours_per_week":8}`
	c, w := newJSONContext(http.MethodPost, "/spaces/1/assignments", body, int64(10))
	c.AddParam("id", "1")
	h.Create(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAssignment_ListBySpace_OK(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	_ = assigns.Create(context.Background(), &domain.Assignment{
		UserID: 5, AcademicSpaceID: 1, ProfessorID: 10,
		RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8,
	})
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/spaces/1/assignments", "", int64(10))
	c.AddParam("id", "1")
	h.ListBySpace(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAssignment_ListMy_OK(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	_ = assigns.Create(context.Background(), &domain.Assignment{
		UserID: 5, AcademicSpaceID: 1, ProfessorID: 10,
		RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8,
	})
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/assignments/me", "", int64(5))
	h.ListMyAssignments(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAssignment_Get_Forbidden(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	_ = assigns.Create(context.Background(), &domain.Assignment{
		UserID: 5, AcademicSpaceID: 1, ProfessorID: 10,
		RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8,
	})
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/assignments/1", "", int64(99))
	c.AddParam("assignmentID", "1")
	h.Get(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestAssignment_ListByProfessor_OK(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	_ = assigns.Create(context.Background(), &domain.Assignment{
		UserID: 5, AcademicSpaceID: 1, ProfessorID: 10,
		RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8,
	})
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/professor/assignments", "", int64(10))
	h.ListByProfessor(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAssignment_ListAllForAdmin_OK(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/admin/assignments", "", nil)
	h.ListAllForAdmin(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAssignment_UpdateByAdmin_OK(t *testing.T) {
	periods, spaces, assigns := seedOpenPeriodAndSpace(t)
	_ = assigns.Create(context.Background(), &domain.Assignment{
		UserID: 5, AcademicSpaceID: 1, ProfessorID: 10,
		RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8,
	})
	svc := appspaces.NewAssignmentService(assigns, spaces, periods, appspaces.NoOpHourRuleChecker{})
	h := NewAssignmentHandler(svc)
	body := `{"role_in_assignment":"graduate_assistant","contracted_hours_per_week":10}`
	c, w := newJSONContext(http.MethodPatch, "/admin/assignments/1", body, nil)
	c.AddParam("assignmentID", "1")
	h.UpdateByAdmin(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func newTaskHandlerForTest(t *testing.T) (*TaskHandler, *handlerTaskRepo) {
	t.Helper()
	repo := newHandlerTaskRepo()
	lookup := &taskAssignmentLookup{byID: map[int64]*domain.Assignment{
		1: {ID: 1, UserID: 1, ProfessorID: 10, AcademicSpaceID: 1, RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8},
	}}
	svc := apptasks.NewTaskService(repo, lookup)
	svc.NowFunc = func() time.Time { return handlerTaskNow }
	return NewTaskHandler(svc), repo
}

func TestTaskHandler_Create_Unauthorized(t *testing.T) {
	h, _ := newTaskHandlerForTest(t)
	body := `{"title":"t","description":"d","status":"Abierto","week_start":"2026-04-06","time_invested":2,"assignment_id":1}`
	c, w := newJSONContext(http.MethodPost, "/tasks", body, nil)
	h.Create(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskHandler_Create_OK(t *testing.T) {
	h, _ := newTaskHandlerForTest(t)
	body := `{"title":"t","description":"d","status":"Abierto","week_start":"2026-04-06","time_invested":2,"assignment_id":1}`
	c, w := newJSONContext(http.MethodPost, "/tasks", body, int64(1))
	h.Create(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_GetAll_OK(t *testing.T) {
	h, repo := newTaskHandlerForTest(t)
	_ = repo.Create(&domain.Task{
		Title: "x", Description: "y", Status: domain.StatusOpen,
		WeekStart: handlerTestMonday, TimeInvested: 2, AssignmentId: 1,
	})
	c, w := newJSONContext(http.MethodGet, "/tasks", "", int64(1))
	h.GetAll(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskHandler_GetByID_NotFound(t *testing.T) {
	h, _ := newTaskHandlerForTest(t)
	c, w := newJSONContext(http.MethodGet, "/tasks/9", "", int64(1))
	c.AddParam("id", "9")
	h.GetByID(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskHandler_Update_InvalidTaskID(t *testing.T) {
	h, _ := newTaskHandlerForTest(t)
	body := `{"title":"t","description":"d","status":"Abierto","time_invested":2}`
	c, w := newJSONContext(http.MethodPut, "/tasks/x", body, int64(1))
	c.AddParam("id", "x")
	h.Update(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskHandler_AdminList_OK(t *testing.T) {
	h, _ := newTaskHandlerForTest(t)
	c, w := newJSONContext(http.MethodGet, "/admin/tasks", "", nil)
	h.AdminList(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskHandler_Delete_OK(t *testing.T) {
	h, repo := newTaskHandlerForTest(t)
	_ = repo.Create(&domain.Task{
		Title: "x", Description: "y", Status: domain.StatusOpen,
		WeekStart: handlerTestMonday, TimeInvested: 2, AssignmentId: 1,
	})
	c, w := newJSONContext(http.MethodDelete, "/tasks/1", "", int64(1))
	c.AddParam("id", "1")
	h.Delete(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_UpdateStatus_OK(t *testing.T) {
	h, repo := newTaskHandlerForTest(t)
	_ = repo.Create(&domain.Task{
		Title: "x", Description: "y", Status: domain.StatusOpen,
		WeekStart: handlerTestMonday, TimeInvested: 2, AssignmentId: 1,
	})
	body := `{"status":"Finalizado"}`
	c, w := newJSONContext(http.MethodPatch, "/tasks/1/status", body, int64(1))
	c.AddParam("id", "1")
	h.UpdateStatus(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_UpdateField_OK(t *testing.T) {
	h, repo := newTaskHandlerForTest(t)
	_ = repo.Create(&domain.Task{
		Title: "x", Description: "y", Status: domain.StatusOpen,
		WeekStart: handlerTestMonday, TimeInvested: 2, AssignmentId: 1,
	})
	body := `{"title":"nuevo"}`
	c, w := newJSONContext(http.MethodPatch, "/tasks/1", body, int64(1))
	c.AddParam("id", "1")
	h.UpdateField(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_UploadAttachment_MissingFile(t *testing.T) {
	h, repo := newTaskHandlerForTest(t)
	_ = repo.Create(&domain.Task{
		Title: "x", Description: "y", Status: domain.StatusOpen,
		WeekStart: handlerTestMonday, TimeInvested: 2, AssignmentId: 1,
	})
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/tasks/1/attachments", nil)
	c.AddParam("id", "1")
	c.Set("authUserID", int64(1))
	h.UploadAttachment(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestTaskHandler_UploadAttachment_OK(t *testing.T) {
	h, repo := newTaskHandlerForTest(t)
	_ = repo.Create(&domain.Task{
		Title: "x", Description: "y", Status: domain.StatusOpen,
		WeekStart: handlerTestMonday, TimeInvested: 2, AssignmentId: 1,
	})
	_ = os.MkdirAll("uploads", 0o755)
	defer os.RemoveAll("uploads")

	var body bytes.Buffer
	mp := multipart.NewWriter(&body)
	part, err := mp.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("hola")); err != nil {
		t.Fatal(err)
	}
	if err := mp.Close(); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/tasks/1/attachments", &body)
	req.Header.Set("Content-Type", mp.FormDataContentType())
	c.Request = req
	c.AddParam("id", "1")
	c.Set("authUserID", int64(1))
	h.UploadAttachment(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func sampleReportAssignments() map[int64][]domain.AssignmentWithUser {
	return map[int64][]domain.AssignmentWithUser{
		10: {{
			Assignment: domain.Assignment{ID: 1, UserID: 3, ProfessorID: 10, RoleInAssignment: domain.RoleMonitor, ContractedHoursPerWeek: 8},
			UserName:   "Juan", UserEmail: "juan@test.com",
		}},
	}
}

func sampleReportTasks() map[string][]domain.Task {
	return map[string][]domain.Task{
		reportTaskKey(1, handlerTestMonday): {{
			ID: 1, Title: "T1", Description: "D", Status: domain.StatusOpen,
			WeekStart: handlerTestMonday, TimeInvested: 3, AssignmentId: 1,
		}},
	}
}

func TestReportHandler_GenerateWeekly_Unauthorized(t *testing.T) {
	svc := appreports.NewReportService(
		newMemoryReportRepo(),
		&reportFakeAssignmentRepo{byProfessor: sampleReportAssignments()},
		&reportFakeTaskRepo{byAssignmentWeek: sampleReportTasks()},
		&reportFakeAI{response: "ok"},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/reports/generate", `{"week_start":"2026-04-06"}`, nil)
	h.GenerateWeekly(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestReportHandler_GenerateWeekly_InvalidWeek(t *testing.T) {
	svc := appreports.NewReportService(
		newMemoryReportRepo(),
		&reportFakeAssignmentRepo{byProfessor: sampleReportAssignments()},
		&reportFakeTaskRepo{byAssignmentWeek: sampleReportTasks()},
		&reportFakeAI{response: "ok"},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/reports/generate", `{"week_start":"2026-04-07"}`, int64(10))
	h.GenerateWeekly(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestReportHandler_GenerateWeekly_OK(t *testing.T) {
	svc := appreports.NewReportService(
		newMemoryReportRepo(),
		&reportFakeAssignmentRepo{byProfessor: sampleReportAssignments()},
		&reportFakeTaskRepo{byAssignmentWeek: sampleReportTasks()},
		&reportFakeAI{response: "Resumen."},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodPost, "/reports/generate", `{"week_start":"2026-04-06"}`, int64(10))
	h.GenerateWeekly(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestReportHandler_List_OK(t *testing.T) {
	repo := newMemoryReportRepo()
	ctx := context.Background()
	_ = repo.Create(ctx, &domain.Report{
		ProfessorID: 10, AssignmentID: 1, UserName: "J", UserEmail: "j@x.co",
		Role: domain.RoleMonitor, WeekStart: handlerTestMonday, FilePath: "/tmp/x.pdf", AISummary: "s",
	})
	svc := appreports.NewReportService(
		repo,
		&reportFakeAssignmentRepo{},
		&reportFakeTaskRepo{},
		&reportFakeAI{},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/reports", "", int64(10))
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestReportHandler_List_ByWeek_InvalidQuery(t *testing.T) {
	svc := appreports.NewReportService(
		newMemoryReportRepo(),
		&reportFakeAssignmentRepo{},
		&reportFakeTaskRepo{},
		&reportFakeAI{},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/reports?week_start=bad", "", int64(10))
	h.List(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestReportHandler_Download_OK(t *testing.T) {
	f, err := os.CreateTemp("", "rep-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	if _, err := f.WriteString("%PDF-1.4"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	defer os.Remove(path)

	repo := newMemoryReportRepo()
	ctx := context.Background()
	_ = repo.Create(ctx, &domain.Report{
		ProfessorID: 10, AssignmentID: 1, UserName: "J", UserEmail: "j@x.co",
		Role: domain.RoleMonitor, WeekStart: handlerTestMonday, FilePath: path, AISummary: "s",
	})
	svc := appreports.NewReportService(
		repo,
		&reportFakeAssignmentRepo{},
		&reportFakeTaskRepo{},
		&reportFakeAI{},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/reports/1/download", "", int64(10))
	c.AddParam("id", "1")
	h.Download(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "pdf") {
		t.Fatalf("content-type=%q", w.Header().Get("Content-Type"))
	}
}

func TestReportHandler_Download_NotFound(t *testing.T) {
	svc := appreports.NewReportService(
		newMemoryReportRepo(),
		&reportFakeAssignmentRepo{},
		&reportFakeTaskRepo{},
		&reportFakeAI{},
		handlerReportPDFStub{},
	)
	h := NewReportHandler(svc)
	c, w := newJSONContext(http.MethodGet, "/reports/99/download", "", int64(10))
	c.AddParam("id", "99")
	h.Download(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestUsers_Post_FirstUser_RequiresAdminRole(t *testing.T) {
	repo := newHandlerUserRepo()
	h := &Users{Admin: &users.AdminService{Users: repo}, JWTSecret: []byte(testJWTSecret)}
	body := `{"name":"A","email":"a@test.com","password":"12345678","roles":["monitor"]}`
	c, w := newJSONContext(http.MethodPost, "/users", body, nil)
	h.Post(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestUsers_Post_FirstUser_OK(t *testing.T) {
	repo := newHandlerUserRepo()
	h := &Users{Admin: &users.AdminService{Users: repo}, JWTSecret: []byte(testJWTSecret)}
	body := `{"name":"A","email":"a@test.com","password":"12345678","roles":["administrador"]}`
	c, w := newJSONContext(http.MethodPost, "/users", body, nil)
	h.Post(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestUsers_Post_SecondUser_WithAdminJWT(t *testing.T) {
	repo := newHandlerUserRepo()
	adminSvc := &users.AdminService{Users: repo}
	ctx := context.Background()
	if _, err := adminSvc.Create(ctx, "Admin", "admin@test.com", "secret1234", []string{domain.RolAdministrador}); err != nil {
		t.Fatal(err)
	}
	login := &auth.LoginService{Users: repo, Secret: []byte(testJWTSecret)}
	res, err := login.Login(ctx, "admin@test.com", "secret1234")
	if err != nil {
		t.Fatal(err)
	}
	h := &Users{Admin: adminSvc, JWTSecret: []byte(testJWTSecret)}
	body := `{"name":"B","email":"b@test.com","password":"12345678","roles":["monitor"]}`
	c, w := newJSONContext(http.MethodPost, "/users", body, nil)
	c.Request.Header.Set("Authorization", "Bearer "+res.Token)
	h.Post(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestUsers_Post_NoBearerWhenUsersExist(t *testing.T) {
	repo := newHandlerUserRepo()
	adminSvc := &users.AdminService{Users: repo}
	ctx := context.Background()
	if _, err := adminSvc.Create(ctx, "Admin", "admin@test.com", "secret1234", []string{domain.RolAdministrador}); err != nil {
		t.Fatal(err)
	}
	h := &Users{Admin: adminSvc, JWTSecret: []byte(testJWTSecret)}
	body := `{"name":"B","email":"b@test.com","password":"12345678","roles":["monitor"]}`
	c, w := newJSONContext(http.MethodPost, "/users", body, nil)
	h.Post(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}

func TestUsers_Post_CountUsersError(t *testing.T) {
	repo := newHandlerUserRepo()
	repo.countErr = errors.New("db down")
	h := &Users{Admin: &users.AdminService{Users: repo}, JWTSecret: []byte(testJWTSecret)}
	body := `{"name":"A","email":"a@test.com","password":"12345678","roles":["administrador"]}`
	c, w := newJSONContext(http.MethodPost, "/users", body, nil)
	h.Post(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestUsers_GetList_OK(t *testing.T) {
	repo := newHandlerUserRepo()
	adminSvc := &users.AdminService{Users: repo}
	ctx := context.Background()
	if _, err := adminSvc.Create(ctx, "A", "a@test.com", "secret1234", []string{domain.RolAdministrador}); err != nil {
		t.Fatal(err)
	}
	h := &Users{Admin: adminSvc, JWTSecret: []byte(testJWTSecret)}
	c, w := newJSONContext(http.MethodGet, "/users", "", nil)
	h.GetList(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d %s", w.Code, w.Body.String())
	}
}
