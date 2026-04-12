package spaces

import (
	"context"
	"fmt"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type AcademicSpaceService struct {
	spaces  domain.AcademicSpaceRepository
	periods domain.AcademicPeriodRepository
}

func NewAcademicSpaceService(
	spaces domain.AcademicSpaceRepository,
	periods domain.AcademicPeriodRepository,
) *AcademicSpaceService {
	return &AcademicSpaceService{spaces: spaces, periods: periods}
}

type CreateSpaceInput struct {
	Name             string
	Type             string
	AcademicPeriodID int64
	ProfessorID      int64
	StartDate        time.Time
	EndDate          time.Time
	Observations     string
}

func (spaceService *AcademicSpaceService) CreateSpace(ctx context.Context, input CreateSpaceInput) (*domain.AcademicSpace, error) {
	period, err := spaceService.periods.FindByID(ctx, input.AcademicPeriodID)
	if err != nil {
		return nil, err
	}
	if !period.IsOpen() {
		return nil, domain.ErrPeriodoCerrado
	}

	space := &domain.AcademicSpace{
		Name:             input.Name,
		Type:             input.Type,
		AcademicPeriodID: input.AcademicPeriodID,
		ProfessorID:      input.ProfessorID,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		Observations:     input.Observations,
		Status:           domain.SpaceStatusActive,
	}

	if err := space.Validate(); err != nil {
		return nil, err
	}

	if err := spaceService.spaces.Create(ctx, space); err != nil {
		return nil, fmt.Errorf("error al crear espacio: %w", err)
	}
	return space, nil
}

func (spaceService *AcademicSpaceService) GetSpace(ctx context.Context, id, professorID int64) (*domain.AcademicSpace, error) {
	space, err := spaceService.spaces.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if space.ProfessorID != professorID {
		return nil, domain.ErrProfesorNoAutorizado
	}
	return space, nil
}

func (spaceService *AcademicSpaceService) ListSpaces(ctx context.Context, professorID int64) ([]domain.AcademicSpace, error) {
	return spaceService.spaces.FindByProfessor(ctx, professorID)
}

func (spaceService *AcademicSpaceService) ListAllSpaces(ctx context.Context) ([]domain.AcademicSpace, error) {
	return spaceService.spaces.ListAll(ctx)
}

func (spaceService *AcademicSpaceService) CloseSpace(ctx context.Context, id, professorID int64) error {
	space, err := spaceService.spaces.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if space.ProfessorID != professorID {
		return domain.ErrProfesorNoAutorizado
	}
	if !space.IsOpen() {
		return domain.ErrEspacioCerrado
	}
	return spaceService.spaces.UpdateStatus(ctx, id, "closed")
}
