package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/jackc/pgx/v5"
)

type AcademicPeriodRepo struct {
	pool *Pool
}

func NewAcademicPeriodRepo(pool *Pool) *AcademicPeriodRepo {
	return &AcademicPeriodRepo{pool: pool}
}

func (r *AcademicPeriodRepo) FindByID(ctx context.Context, id int64) (*domain.AcademicPeriod, error) {
	const q = `
		SELECT id, code, start_date, end_date, status
		FROM academic_periods
		WHERE id = $1`

	row := r.pool.inner.QueryRow(ctx, q, id)
	var p domain.AcademicPeriod
	err := row.Scan(&p.ID, &p.Code, &p.StartDate, &p.EndDate, &p.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPeriodoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("academic_period FindByID: %w", err)
	}
	return &p, nil
}

func (r *AcademicPeriodRepo) Create(ctx context.Context, p *domain.AcademicPeriod) error {
	const q = `
		INSERT INTO academic_periods (code, start_date, end_date, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	row := r.pool.inner.QueryRow(ctx, q, p.Code, p.StartDate, p.EndDate, p.Status)
	if err := row.Scan(&p.ID, &p.CreatedAt); err != nil {
		return fmt.Errorf("academic_period Create: %w", err)
	}
	return nil
}

func (r *AcademicPeriodRepo) List(ctx context.Context) ([]domain.AcademicPeriod, error) {
	const q = `
		SELECT id, code, start_date, end_date, status
		FROM academic_periods
		ORDER BY created_at DESC`

	rows, err := r.pool.inner.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("academic_period List: %w", err)
	}
	defer rows.Close()

	var result []domain.AcademicPeriod
	for rows.Next() {
		var p domain.AcademicPeriod
		if err := rows.Scan(&p.ID, &p.Code, &p.StartDate, &p.EndDate, &p.Status); err != nil {
			return nil, fmt.Errorf("academic_period scan: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *AcademicPeriodRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	const q = `UPDATE academic_periods SET status = $1 WHERE id = $2`
	tag, err := r.pool.inner.Exec(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("academic_period UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPeriodoNoEncontrado
	}
	return nil
}