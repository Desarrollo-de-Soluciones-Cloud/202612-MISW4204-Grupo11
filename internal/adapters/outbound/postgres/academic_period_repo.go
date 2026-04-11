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

func (periodRepo *AcademicPeriodRepo) FindByID(ctx context.Context, id int64) (*domain.AcademicPeriod, error) {
	const query = `
		SELECT id, code, start_date, end_date, status
		FROM academic_periods
		WHERE id = $1`

	row := periodRepo.pool.inner.QueryRow(ctx, query, id)
	var period domain.AcademicPeriod
	err := row.Scan(&period.ID, &period.Code, &period.StartDate, &period.EndDate, &period.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPeriodoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("academic_period FindByID: %w", err)
	}
	return &period, nil
}

func (periodRepo *AcademicPeriodRepo) Create(ctx context.Context, period *domain.AcademicPeriod) error {
	const query = `
		INSERT INTO academic_periods (code, start_date, end_date, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	row := periodRepo.pool.inner.QueryRow(ctx, query, period.Code, period.StartDate, period.EndDate, period.Status)
	if err := row.Scan(&period.ID, &period.CreatedAt); err != nil {
		return fmt.Errorf("academic_period Create: %w", err)
	}
	return nil
}

func (periodRepo *AcademicPeriodRepo) List(ctx context.Context) ([]domain.AcademicPeriod, error) {
	const query = `
		SELECT id, code, start_date, end_date, status
		FROM academic_periods
		ORDER BY created_at DESC`

	rows, err := periodRepo.pool.inner.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("academic_period List: %w", err)
	}
	defer rows.Close()

	var result []domain.AcademicPeriod
	for rows.Next() {
		var period domain.AcademicPeriod
		if err := rows.Scan(&period.ID, &period.Code, &period.StartDate, &period.EndDate, &period.Status); err != nil {
			return nil, fmt.Errorf("academic_period scan: %w", err)
		}
		result = append(result, period)
	}
	return result, rows.Err()
}

func (periodRepo *AcademicPeriodRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	const query = `UPDATE academic_periods SET status = $1 WHERE id = $2`
	tag, err := periodRepo.pool.inner.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("academic_period UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPeriodoNoEncontrado
	}
	return nil
}
