package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/jackc/pgx/v5"
)

type AcademicSpaceRepo struct {
	pool *Pool
}

func NewAcademicSpaceRepo(pool *Pool) *AcademicSpaceRepo {
	return &AcademicSpaceRepo{pool: pool}
}

func (r *AcademicSpaceRepo) Create(ctx context.Context, s *domain.AcademicSpace) error {
	const q = `
		INSERT INTO academic_spaces
			(name, type, academic_period_id, professor_id, start_date, end_date, observations, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	row := r.pool.inner.QueryRow(ctx, q,
		s.Name, s.Type, s.AcademicPeriodID, s.ProfessorID,
		s.StartDate, s.EndDate, s.Observations, s.Status,
	)
	if err := row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return fmt.Errorf("academic_space Create: %w", err)
	}
	return nil
}

func (r *AcademicSpaceRepo) FindByID(ctx context.Context, id int64) (*domain.AcademicSpace, error) {
	const q = `
		SELECT id, name, type, academic_period_id, professor_id,
		start_date, end_date, observations, status, created_at, updated_at
		FROM academic_spaces
		WHERE id = $1`

	row := r.pool.inner.QueryRow(ctx, q, id)
	s, err := scanSpace(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEspacioNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("academic_space FindByID: %w", err)
	}
	return s, nil
}

func (r *AcademicSpaceRepo) FindByProfessor(ctx context.Context, professorID int64) ([]domain.AcademicSpace, error) {
	const q = `
		SELECT id, name, type, academic_period_id, professor_id,
		start_date, end_date, observations, status, created_at, updated_at
		FROM academic_spaces
		WHERE professor_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.inner.Query(ctx, q, professorID)
	if err != nil {
		return nil, fmt.Errorf("academic_space FindByProfessor: %w", err)
	}
	defer rows.Close()

	var result []domain.AcademicSpace
	for rows.Next() {
		s, err := scanSpace(rows)
		if err != nil {
			return nil, fmt.Errorf("academic_space scan: %w", err)
		}
		result = append(result, *s)
	}
	return result, rows.Err()
}

func (r *AcademicSpaceRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	const q = `UPDATE academic_spaces SET status = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.pool.inner.Exec(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("academic_space UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEspacioNoEncontrado
	}
	return nil
}

// scanSpace es un helper para no repetir el Scan en cada método.
type scannable interface {
	Scan(dest ...any) error
}

func scanSpace(row scannable) (*domain.AcademicSpace, error) {
	var s domain.AcademicSpace
	err := row.Scan(
		&s.ID, &s.Name, &s.Type, &s.AcademicPeriodID, &s.ProfessorID,
		&s.StartDate, &s.EndDate, &s.Observations, &s.Status,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
