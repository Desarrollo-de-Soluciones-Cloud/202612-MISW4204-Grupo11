package domain

import "context"

type AcademicPeriodRepository interface {
	FindByID(ctx context.Context, id int64) (*AcademicPeriod, error)
	Create(ctx context.Context, p *AcademicPeriod) error
	List(ctx context.Context) ([]AcademicPeriod, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type AcademicSpaceRepository interface {
	Create(ctx context.Context, space *AcademicSpace) error
	FindByID(ctx context.Context, id int64) (*AcademicSpace, error)
	FindByProfessor(ctx context.Context, professorID int64) ([]AcademicSpace, error)
	ListAll(ctx context.Context) ([]AcademicSpace, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *Assignment) error
	FindByID(ctx context.Context, id int64) (*Assignment, error)
	FindBySpace(ctx context.Context, spaceID int64) ([]Assignment, error)
	FindByUser(ctx context.Context, userID int64) ([]Assignment, error)
	ExistsByUserSpaceRole(ctx context.Context, userID, spaceID int64, role string) (bool, error)
	FindActiveByUserAndRole(ctx context.Context, userID int64, role string) ([]Assignment, error)
	FindByProfessorWithUser(ctx context.Context, professorID int64) ([]AssignmentWithUser, error)
	ListAll(ctx context.Context) ([]Assignment, error)
	Update(ctx context.Context, assignment *Assignment) error
}
