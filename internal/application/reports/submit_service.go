package reports

import (
	"context"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/google/uuid"
)

type WeeklyReportSubmitter interface {
	SubmitWeeklyReportJob(ctx context.Context, professorID int64, weekStart time.Time) (ports.WeeklyReportJob, error)
}

type SubmitService struct {
	publisher ports.ReportJobPublisher
	nowFunc   func() time.Time
	newIDFunc func() string
}

func NewSubmitService(publisher ports.ReportJobPublisher) *SubmitService {
	return &SubmitService{
		publisher: publisher,
		nowFunc:   time.Now,
		newIDFunc: uuid.NewString,
	}
}

func (s *SubmitService) SubmitWeeklyReportJob(ctx context.Context, professorID int64, weekStart time.Time) (ports.WeeklyReportJob, error) {
	if err := domain.ValidateWeekStart(weekStart); err != nil {
		return ports.WeeklyReportJob{}, err
	}

	job := ports.WeeklyReportJob{
		RequestID:   s.newIDFunc(),
		ProfessorID: professorID,
		WeekStart:   weekStart.Format(time.DateOnly),
		RequestedAt: s.nowFunc().UTC(),
	}

	if err := s.publisher.PublishWeeklyReportJob(ctx, job); err != nil {
		return ports.WeeklyReportJob{}, err
	}

	return job, nil
}
