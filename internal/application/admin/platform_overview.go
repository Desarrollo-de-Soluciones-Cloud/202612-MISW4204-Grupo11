package admin

import (
	"context"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type TaskLister interface {
	ListAll(ctx context.Context) ([]domain.Task, error)
}

type PlatformOverviewService struct {
	Users       ports.UserRepository
	Periods     domain.AcademicPeriodRepository
	Spaces      domain.AcademicSpaceRepository
	Assignments domain.AssignmentRepository
	Tasks       TaskLister
}

func NewPlatformOverviewService(
	users ports.UserRepository,
	periods domain.AcademicPeriodRepository,
	spaces domain.AcademicSpaceRepository,
	assignments domain.AssignmentRepository,
	tasks TaskLister,
) *PlatformOverviewService {
	return &PlatformOverviewService{
		Users:       users,
		Periods:     periods,
		Spaces:      spaces,
		Assignments: assignments,
		Tasks:       tasks,
	}
}

type PlatformOverview struct {
	Users           []domain.User           `json:"users"`
	AcademicPeriods []domain.AcademicPeriod `json:"academic_periods"`
	AcademicSpaces  []domain.AcademicSpace  `json:"academic_spaces"`
	Assignments     []domain.Assignment     `json:"assignments"`
	Tasks           []domain.Task           `json:"tasks"`
}

func (service *PlatformOverviewService) GetOverview(ctx context.Context) (PlatformOverview, error) {
	var out PlatformOverview

	userList, err := service.Users.ListUsers(ctx)
	if err != nil {
		return out, fmt.Errorf("list users: %w", err)
	}
	out.Users = userList

	periods, err := service.Periods.List(ctx)
	if err != nil {
		return out, fmt.Errorf("list periods: %w", err)
	}
	out.AcademicPeriods = periods

	spaces, err := service.Spaces.ListAll(ctx)
	if err != nil {
		return out, fmt.Errorf("list spaces: %w", err)
	}
	out.AcademicSpaces = spaces

	assignments, err := service.Assignments.ListAll(ctx)
	if err != nil {
		return out, fmt.Errorf("list assignments: %w", err)
	}
	out.Assignments = assignments

	tasks, err := service.Tasks.ListAll(ctx)
	if err != nil {
		return out, fmt.Errorf("list tasks: %w", err)
	}
	out.Tasks = tasks

	return out, nil
}
