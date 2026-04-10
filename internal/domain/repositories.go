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
	UpdateStatus(ctx context.Context, id int64, status string) error
}

