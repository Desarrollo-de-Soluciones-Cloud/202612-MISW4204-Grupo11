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
