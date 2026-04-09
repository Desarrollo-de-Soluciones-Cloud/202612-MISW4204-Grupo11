package spaces

import (
	"context"
	"fmt"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

//  implementa los casos de uso para gestionar espacios academicos
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

//  agrupa los datos necesarios para crear un espacio
type CreateSpaceInput struct {
	Name             string
	Type             string
	AcademicPeriodID int64
	ProfessorID      int64
	StartDate        time.Time
	EndDate          time.Time
	Observations     string
}

// crea un curso o proyecto 
func (s *AcademicSpaceService) CreateSpace(ctx context.Context, in CreateSpaceInput) (*domain.AcademicSpace, error) {
	// el período debe existir y estar activo.
	period, err := s.periods.FindByID(ctx, in.AcademicPeriodID)
	if err != nil {
		return nil, err
	}
	if !period.IsOpen() {
		return nil, domain.ErrPeriodoCerrado
	}

	space := &domain.AcademicSpace{
		Name:             in.Name,
		Type:             in.Type,
		AcademicPeriodID: in.AcademicPeriodID,
		ProfessorID:      in.ProfessorID,
		StartDate:        in.StartDate,
		EndDate:          in.EndDate,
		Observations:     in.Observations,
		Status:           domain.SpaceStatusActive,
	}

	// Validación de dominio (tipo válido, fechas coherentes).
	if err := space.Validate(); err != nil {
		return nil, err
	}

	if err := s.spaces.Create(ctx, space); err != nil {
		return nil, fmt.Errorf("error al crear espacio: %w", err)
	}
	return space, nil
}

// GetSpace devuelve un espacio si pertenece al profesor
func (s *AcademicSpaceService) GetSpace(ctx context.Context, id, professorID int64) (*domain.AcademicSpace, error) {
	space, err := s.spaces.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if space.ProfessorID != professorID {
		return nil, domain.ErrProfesorNoAutorizado
	}
	return space, nil
}

// ListSpaces devuelve todos los espacios del profesor
func (s *AcademicSpaceService) ListSpaces(ctx context.Context, professorID int64) ([]domain.AcademicSpace, error) {
	return s.spaces.FindByProfessor(ctx, professorID)
}

// CloseSpace cambia el estado del espacio a cerrado
func (s *AcademicSpaceService) CloseSpace(ctx context.Context, id, professorID int64) error {
	space, err := s.spaces.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if space.ProfessorID != professorID {
		return domain.ErrProfesorNoAutorizado
	}
	if !space.IsActive() {
		return domain.ErrEspacioCerrado
	}
	return s.spaces.UpdateStatus(ctx, id, domain.SpaceStatusClosed)
}