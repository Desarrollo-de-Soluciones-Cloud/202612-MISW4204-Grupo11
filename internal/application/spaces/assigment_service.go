package spaces

import (
	"context"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type HourRuleChecker interface {
	CheckNewAssignment(ctx context.Context, userID int64, role string, hoursPerWeek int) error
}

// stub temporal para no bloquear el desarrollo de otras partes mientras se define la lógica de validación de horas
type NoOpHourRuleChecker struct{}

func (NoOpHourRuleChecker) CheckNewAssignment(_ context.Context, _ int64, _ string, _ int) error {
	return nil
}

type AssignmentService struct {
	assignmentRepo domain.AssignmentRepository
	spaceRepo      domain.AcademicSpaceRepository
	periodRepo     domain.AcademicPeriodRepository
	hourChecker    HourRuleChecker
}

func NewAssignmentService(
	assignmentRepo domain.AssignmentRepository,
	spaceRepo domain.AcademicSpaceRepository,
	periodRepo domain.AcademicPeriodRepository,
	hourChecker HourRuleChecker,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		spaceRepo:      spaceRepo,
		periodRepo:     periodRepo,
		hourChecker:    hourChecker,
	}
}

type CreateAssignmentInput struct {
	UserID                 int64
	AcademicSpaceID        int64
	ProfessorID            int64
	RoleInAssignment       string
	ContractedHoursPerWeek int
}

func (service *AssignmentService) CreateAssignment(ctx context.Context, input CreateAssignmentInput) (*domain.Assignment, error) {

	space, err := service.spaceRepo.FindByID(ctx, input.AcademicSpaceID)
	if err != nil {
		return nil, err
	}
	if space.ProfessorID != input.ProfessorID {
		return nil, domain.ErrProfesorNoAutorizado
	}

	if !space.IsActive() {
		return nil, domain.ErrEspacioCerradoVinculacion
	}

	period, err := service.periodRepo.FindByID(ctx, space.AcademicPeriodID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener período académico: %w", err)
	}

	if !period.IsOpen() {
		return nil, domain.ErrPeriodoCerradoVinculacion
	}

	if space.StartDate.Before(period.StartDate) || space.EndDate.After(period.EndDate) {
		return nil, domain.ErrFechasEspacioFueraDelPeriodo
	}

	assignment := &domain.Assignment{
		UserID:                 input.UserID,
		AcademicSpaceID:        input.AcademicSpaceID,
		ProfessorID:            input.ProfessorID,
		RoleInAssignment:       input.RoleInAssignment,
		ContractedHoursPerWeek: input.ContractedHoursPerWeek,
	}
	if err := assignment.Validate(); err != nil {
		return nil, err
	}

	exists, err := service.assignmentRepo.ExistsByUserSpaceRole(ctx, input.UserID, input.AcademicSpaceID, input.RoleInAssignment)
	if err != nil {
		return nil, fmt.Errorf("error verificando duplicado: %w", err)
	}
	if exists {
		return nil, domain.ErrUsuarioYaVinculado
	}

	if err := service.hourChecker.CheckNewAssignment(ctx, input.UserID, input.RoleInAssignment, input.ContractedHoursPerWeek); err != nil {
		return nil, err
	}

	listUser, err := service.assignmentRepo.FindByUser(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("error consultando vinculaciones del usuario: %w", err)
	}

	listUser = append(listUser, *assignment)

	resultAboveHours := domain.CheckMaxHoursPerRole(listUser)
	resultNumberOfClases := domain.LimitClasesPerUser(listUser)
	resultRuleCombined := domain.Validar40PercentOfMonitorHours(listUser)

	if resultNumberOfClases {
		return nil, fmt.Errorf("El usuario ya tiene mas e 3 monitorias, no se puede agregar otra monitoria mas.")
	}

	if resultAboveHours {
		return nil, fmt.Errorf("El usuario ha pasado el limite de horas para su role " + input.RoleInAssignment)
	}

	if resultRuleCombined {
		return nil, fmt.Errorf("El no puede asignarse mas horas de monitoria. Si se agrega esta vinculacion, se estaria pasando el limite de 40 porciente de las horas contratadas como " + input.RoleInAssignment)
	}

	if err := service.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, fmt.Errorf("error al crear vinculación: %w", err)
	}

	return assignment, nil
}

func (service *AssignmentService) GetAssignment(ctx context.Context, id, professorID int64) (*domain.Assignment, error) {
	assignment, err := service.assignmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if assignment.ProfessorID != professorID {
		return nil, domain.ErrProfesorNoAutorizado
	}
	return assignment, nil
}

func (service *AssignmentService) ListAssignmentsBySpace(ctx context.Context, spaceID, professorID int64) ([]domain.Assignment, error) {
	space, err := service.spaceRepo.FindByID(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	if space.ProfessorID != professorID {
		return nil, domain.ErrProfesorNoAutorizado
	}
	return service.assignmentRepo.FindBySpace(ctx, spaceID)
}

func (service *AssignmentService) ListAssignmentsByUser(ctx context.Context, userID int64) ([]domain.Assignment, error) {
	return service.assignmentRepo.FindByUser(ctx, userID)
}

func (service *AssignmentService) ListAssignmentsByProfessor(ctx context.Context, professorID int64) ([]domain.AssignmentWithUser, error) {
	return service.assignmentRepo.FindByProfessorWithUser(ctx, professorID)
}

type UpdateAssignmentInput struct {
	RoleInAssignment       string
	ContractedHoursPerWeek int
}

func (service *AssignmentService) UpdateAssignmentByAdmin(ctx context.Context, assignmentID int64, input UpdateAssignmentInput) (*domain.Assignment, error) {
	assignment, err := service.assignmentRepo.FindByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	assignment.RoleInAssignment = input.RoleInAssignment
	assignment.ContractedHoursPerWeek = input.ContractedHoursPerWeek

	if err := assignment.Validate(); err != nil {
		return nil, err
	}

	if err := service.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}
