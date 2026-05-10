package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	appreports "github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/reports"
)

type submitterStub struct {
	err error
}

func (s submitterStub) SubmitWeeklyReportJob(_ context.Context, professorID int64, weekStart time.Time) (ports.WeeklyReportJob, error) {
	if s.err != nil {
		return ports.WeeklyReportJob{}, s.err
	}
	return ports.WeeklyReportJob{
		RequestID:   "req-async",
		ProfessorID: professorID,
		WeekStart:   weekStart.Format(time.DateOnly),
		RequestedAt: time.Now().UTC(),
	}, nil
}

type submitterSpy struct {
	called      int
	lastUserID  int64
	lastWeek    time.Time
	responseJob ports.WeeklyReportJob
}

func (s *submitterSpy) SubmitWeeklyReportJob(_ context.Context, professorID int64, weekStart time.Time) (ports.WeeklyReportJob, error) {
	s.called++
	s.lastUserID = professorID
	s.lastWeek = weekStart
	if s.responseJob.RequestID == "" {
		s.responseJob = ports.WeeklyReportJob{RequestID: "req-spy"}
	}
	return s.responseJob, nil
}

func TestReportHandler_GenerateWeekly_QueuesJobWhenSubmitterConfigured(t *testing.T) {
	svc := appreports.NewReportService(newMemoryReportRepo(), &reportFakeAssignmentRepo{}, &reportFakeTaskRepo{}, &reportFakeAI{}, handlerReportPDFStub{})
	h := NewReportHandler(svc, submitterStub{})

	c, w := newJSONContext(http.MethodPost, "/reports/weekly", `{"week_start":"2026-04-06"}`, int64(10))
	h.GenerateWeekly(c)

	if w.Code != http.StatusAccepted {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestReportHandler_GenerateWeekly_QueueError(t *testing.T) {
	svc := appreports.NewReportService(newMemoryReportRepo(), &reportFakeAssignmentRepo{}, &reportFakeTaskRepo{}, &reportFakeAI{}, handlerReportPDFStub{})
	h := NewReportHandler(svc, submitterStub{err: errors.New("broker unavailable")})

	c, w := newJSONContext(http.MethodPost, "/reports/weekly", `{"week_start":"2026-04-06"}`, int64(10))
	h.GenerateWeekly(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestReportHandler_GenerateWeekly_QueuesJob_ResponsePayload(t *testing.T) {
	spy := &submitterSpy{responseJob: ports.WeeklyReportJob{RequestID: "req-123"}}
	svc := appreports.NewReportService(newMemoryReportRepo(), &reportFakeAssignmentRepo{}, &reportFakeTaskRepo{}, &reportFakeAI{}, handlerReportPDFStub{})
	h := NewReportHandler(svc, spy)

	c, w := newJSONContext(http.MethodPost, "/reports/weekly", `{"week_start":"2026-04-06"}`, int64(10))
	h.GenerateWeekly(c)

	if w.Code != http.StatusAccepted {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	if spy.called != 1 {
		t.Fatalf("submitter should be called once, got %d", spy.called)
	}
	if spy.lastUserID != 10 {
		t.Fatalf("unexpected professor id: %d", spy.lastUserID)
	}

	var payload map[string]string
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["request_id"] != "req-123" {
		t.Fatalf("unexpected request_id: %s", payload["request_id"])
	}
	if payload["status"] != "queued" {
		t.Fatalf("unexpected status: %s", payload["status"])
	}
}

func TestReportHandler_GenerateWeekly_WithSubmitter_UnauthorizedDoesNotQueue(t *testing.T) {
	spy := &submitterSpy{}
	svc := appreports.NewReportService(newMemoryReportRepo(), &reportFakeAssignmentRepo{}, &reportFakeTaskRepo{}, &reportFakeAI{}, handlerReportPDFStub{})
	h := NewReportHandler(svc, spy)

	c, w := newJSONContext(http.MethodPost, "/reports/weekly", `{"week_start":"2026-04-06"}`, nil)
	h.GenerateWeekly(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	if spy.called != 0 {
		t.Fatalf("submitter must not be called on unauthorized request")
	}
}

func TestReportHandler_GenerateWeekly_WithSubmitter_InvalidWeekDoesNotQueue(t *testing.T) {
	spy := &submitterSpy{}
	svc := appreports.NewReportService(newMemoryReportRepo(), &reportFakeAssignmentRepo{}, &reportFakeTaskRepo{}, &reportFakeAI{}, handlerReportPDFStub{})
	h := NewReportHandler(svc, spy)

	c, w := newJSONContext(http.MethodPost, "/reports/weekly", `{"week_start":"not-a-date"}`, int64(10))
	h.GenerateWeekly(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	if spy.called != 0 {
		t.Fatalf("submitter must not be called when payload is invalid")
	}
}
