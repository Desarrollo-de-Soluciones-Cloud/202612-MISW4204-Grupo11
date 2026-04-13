package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/jackc/pgx/v5"
)

type ReportRepo struct {
	pool *Pool
}

func NewReportRepo(pool *Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

func (r *ReportRepo) Create(ctx context.Context, report *domain.Report) error {
	const query = `
		INSERT INTO reports
			(professor_id, assignment_id, user_name, user_email, role, week_start, file_path, ai_summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (assignment_id, week_start) DO UPDATE SET
			file_path = EXCLUDED.file_path,
			ai_summary = EXCLUDED.ai_summary,
			created_at = NOW()
		RETURNING id, created_at`

	row := r.pool.inner.QueryRow(ctx, query,
		report.ProfessorID, report.AssignmentID, report.UserName, report.UserEmail,
		report.Role, report.WeekStart, report.FilePath, report.AISummary,
	)
	if err := row.Scan(&report.ID, &report.CreatedAt); err != nil {
		return fmt.Errorf("report Create: %w", err)
	}
	return nil
}

func (r *ReportRepo) FindByID(ctx context.Context, id int64) (*domain.Report, error) {
	const query = `
		SELECT id, professor_id, assignment_id, user_name, user_email, role,
			week_start, file_path, ai_summary, created_at
		FROM reports WHERE id = $1`

	row := r.pool.inner.QueryRow(ctx, query, id)
	report, err := scanReport(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrReporteNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("report FindByID: %w", err)
	}
	return report, nil
}

func (r *ReportRepo) FindByProfessorAndWeek(ctx context.Context, professorID int64, weekStart time.Time) ([]domain.Report, error) {
	const query = `
		SELECT id, professor_id, assignment_id, user_name, user_email, role,
			week_start, file_path, ai_summary, created_at
		FROM reports WHERE professor_id = $1 AND week_start = $2
		ORDER BY user_name`

	rows, err := r.pool.inner.Query(ctx, query, professorID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("report FindByProfessorAndWeek: %w", err)
	}
	defer rows.Close()

	return scanReports(rows)
}

func (r *ReportRepo) FindByProfessor(ctx context.Context, professorID int64) ([]domain.Report, error) {
	const query = `
		SELECT id, professor_id, assignment_id, user_name, user_email, role,
			week_start, file_path, ai_summary, created_at
		FROM reports WHERE professor_id = $1
		ORDER BY week_start DESC, user_name`

	rows, err := r.pool.inner.Query(ctx, query, professorID)
	if err != nil {
		return nil, fmt.Errorf("report FindByProfessor: %w", err)
	}
	defer rows.Close()

	return scanReports(rows)
}

func scanReport(row interface{ Scan(dest ...any) error }) (*domain.Report, error) {
	var r domain.Report
	err := row.Scan(
		&r.ID, &r.ProfessorID, &r.AssignmentID, &r.UserName, &r.UserEmail,
		&r.Role, &r.WeekStart, &r.FilePath, &r.AISummary, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func scanReports(rows pgx.Rows) ([]domain.Report, error) {
	var reports []domain.Report
	for rows.Next() {
		r, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("report scan: %w", err)
		}
		reports = append(reports, *r)
	}
	return reports, rows.Err()
}
