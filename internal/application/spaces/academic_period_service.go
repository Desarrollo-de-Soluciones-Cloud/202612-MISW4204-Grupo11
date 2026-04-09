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

func (periodService *AcademicPeriodService) CreatePeriod(ctx context.Context, in CreatePeriodInput) (*domain.AcademicPeriod, error) {
	if !in.EndDate.After(in.StartDate) {
		return nil, domain.ErrFechasCierreInvalidas
	}
	period := &domain.AcademicPeriod{
		Code:      in.Code,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Status:    "active",
	}
	if err := periodService.periods.Create(ctx, period); err != nil {
		return nil, fmt.Errorf("error al crear período: %w", err)
	}
	return period, nil
}

func (periodService *AcademicPeriodService) ListPeriods(ctx context.Context) ([]domain.AcademicPeriod, error) {
	return periodService.periods.List(ctx)
}

func (periodService *AcademicPeriodService) ClosePeriod(ctx context.Context, id int64) error {
	period, err := periodService.periods.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !period.IsOpen() {
		return domain.ErrPeriodoCerrado
	}
	return periodService.periods.UpdateStatus(ctx, id, "closed")
}
