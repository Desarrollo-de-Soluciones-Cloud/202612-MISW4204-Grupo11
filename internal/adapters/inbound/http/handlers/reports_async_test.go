package handlers

import (
	"context"
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
