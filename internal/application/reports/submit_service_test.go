package reports

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type publishStub struct {
	lastJob ports.WeeklyReportJob
	err     error
}

func (p *publishStub) PublishWeeklyReportJob(_ context.Context, job ports.WeeklyReportJob) error {
	if p.err != nil {
		return p.err
	}
	p.lastJob = job
	return nil
}

func TestSubmitWeeklyReportJob_OK(t *testing.T) {
	pub := &publishStub{}
	svc := NewSubmitService(pub)
	svc.nowFunc = func() time.Time { return time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC) }
	svc.newIDFunc = func() string { return "req-123" }

	weekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	job, err := svc.SubmitWeeklyReportJob(context.Background(), 10, weekStart)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if job.RequestID != "req-123" || job.ProfessorID != 10 || job.WeekStart != "2026-04-06" {
		t.Fatalf("unexpected job %+v", job)
	}
	if pub.lastJob.RequestID != "req-123" {
		t.Fatalf("publisher did not receive job: %+v", pub.lastJob)
	}
}

func TestSubmitWeeklyReportJob_InvalidWeekStart(t *testing.T) {
	pub := &publishStub{}
	svc := NewSubmitService(pub)

	invalidWeek := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)
	_, err := svc.SubmitWeeklyReportJob(context.Background(), 10, invalidWeek)
	if !errors.Is(err, domain.ErrSemanaInicioNoEsLunes) {
		t.Fatalf("expected ErrSemanaInicioNoEsLunes, got %v", err)
	}
}

func TestSubmitWeeklyReportJob_PublishError(t *testing.T) {
	pub := &publishStub{err: errors.New("broker down")}
	svc := NewSubmitService(pub)

	weekStart := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	_, err := svc.SubmitWeeklyReportJob(context.Background(), 10, weekStart)
	if err == nil || err.Error() != "broker down" {
		t.Fatalf("expected publish error, got %v", err)
	}
}
