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

func (spaceRepo *AcademicSpaceRepo) Create(ctx context.Context, space *domain.AcademicSpace) error {
	const query = `
		INSERT INTO academic_spaces
			(name, type, academic_period_id, professor_id, start_date, end_date, observations, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	row := spaceRepo.pool.inner.QueryRow(ctx, query,
		space.Name, space.Type, space.AcademicPeriodID, space.ProfessorID,
		space.StartDate, space.EndDate, space.Observations, space.Status,
	)
	if err := row.Scan(&space.ID, &space.CreatedAt, &space.UpdatedAt); err != nil {
		return fmt.Errorf("academic_space Create: %w", err)
	}
	return nil
}

func (spaceRepo *AcademicSpaceRepo) FindByID(ctx context.Context, id int64) (*domain.AcademicSpace, error) {
	const query = `
		SELECT id, name, type, academic_period_id, professor_id,
		start_date, end_date, observations, status, created_at, updated_at
		FROM academic_spaces
		WHERE id = $1`

	row := spaceRepo.pool.inner.QueryRow(ctx, query, id)
	space, err := scanSpace(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEspacioNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("academic_space FindByID: %w", err)
	}
	return space, nil
}

func (spaceRepo *AcademicSpaceRepo) FindByProfessor(ctx context.Context, professorID int64) ([]domain.AcademicSpace, error) {
	const query = `
		SELECT id, name, type, academic_period_id, professor_id,
		start_date, end_date, observations, status, created_at, updated_at
		FROM academic_spaces
		WHERE professor_id = $1
		ORDER BY created_at DESC`

	rows, err := spaceRepo.pool.inner.Query(ctx, query, professorID)
	if err != nil {
		return nil, fmt.Errorf("academic_space FindByProfessor: %w", err)
	}
	defer rows.Close()

	var result []domain.AcademicSpace
	for rows.Next() {
		space, err := scanSpace(rows)
		if err != nil {
			return nil, fmt.Errorf("academic_space scan: %w", err)
		}
		result = append(result, *space)
	}
	return result, rows.Err()
}

func (spaceRepo *AcademicSpaceRepo) ListAll(ctx context.Context) ([]domain.AcademicSpace, error) {
	const query = `
		SELECT id, name, type, academic_period_id, professor_id,
		start_date, end_date, observations, status, created_at, updated_at
		FROM academic_spaces
		ORDER BY id ASC`

	rows, err := spaceRepo.pool.inner.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("academic_space ListAll: %w", err)
	}
	defer rows.Close()

	var result []domain.AcademicSpace
	for rows.Next() {
		space, err := scanSpace(rows)
		if err != nil {
			return nil, fmt.Errorf("academic_space scan: %w", err)
		}
		result = append(result, *space)
	}
	return result, rows.Err()
}

func (spaceRepo *AcademicSpaceRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	const query = `UPDATE academic_spaces SET status = $1, updated_at = NOW() WHERE id = $2`
	tag, err := spaceRepo.pool.inner.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("academic_space UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEspacioNoEncontrado
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanSpace(row scannable) (*domain.AcademicSpace, error) {
	var space domain.AcademicSpace
	err := row.Scan(
		&space.ID, &space.Name, &space.Type, &space.AcademicPeriodID, &space.ProfessorID,
		&space.StartDate, &space.EndDate, &space.Observations, &space.Status,
		&space.CreatedAt, &space.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &space, nil
}
