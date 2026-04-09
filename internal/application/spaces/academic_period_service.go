package spaces

import (
	"context"
	"fmt"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type AcademicPeriodService struct {
	periods domain.AcademicPeriodRepository
}

func NewAcademicPeriodService(periods domain.AcademicPeriodRepository) *AcademicPeriodService {
	return &AcademicPeriodService{periods: periods}
}

type CreatePeriodInput struct {
	Code      string
	StartDate time.Time
	EndDate   time.Time
}

func (s *AcademicPeriodService) CreatePeriod(ctx context.Context, in CreatePeriodInput) (*domain.AcademicPeriod, error) {
	if !in.EndDate.After(in.StartDate) {
		return nil, domain.ErrFechasCierreInvalidas
	}
	p := &domain.AcademicPeriod{
		Code:      in.Code,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Status:    "active",
	}
	if err := s.periods.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("error al crear período: %w", err)
	}
	return p, nil
}

func (s *AcademicPeriodService) ListPeriods(ctx context.Context) ([]domain.AcademicPeriod, error) {
	return s.periods.List(ctx)
}

func (s *AcademicPeriodService) ClosePeriod(ctx context.Context, id int64) error {
	p, err := s.periods.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !p.IsOpen() {
		return domain.ErrPeriodoCerrado
	}
	return s.periods.UpdateStatus(ctx, id, "closed")
}