package domain

import "context"

//  define las operaciones de persistencia para períodos académicos.
type AcademicPeriodRepository interface {
	FindByID(ctx context.Context, id int64) (*AcademicPeriod, error)
}

//  define las operaciones de persistencia para espacios académicos.
type AcademicSpaceRepository interface {
	Create(ctx context.Context, space *AcademicSpace) error
	FindByID(ctx context.Context, id int64) (*AcademicSpace, error)
	FindByProfessor(ctx context.Context, professorID int64) ([]AcademicSpace, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

