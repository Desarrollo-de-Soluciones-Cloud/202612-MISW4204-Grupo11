package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/jackc/pgx/v5"
)

type AssignmentRepo struct {
	pool *Pool
}

func NewAssignmentRepo(pool *Pool) *AssignmentRepo {
	return &AssignmentRepo{pool: pool}
}

func (assignmentRepo *AssignmentRepo) Create(ctx context.Context, assignment *domain.Assignment) error {
	const query = `
		INSERT INTO assignments
			(user_id, academic_space_id, professor_id, role_in_assignment, contracted_hours_per_week)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	row := assignmentRepo.pool.inner.QueryRow(ctx, query,
		assignment.UserID, assignment.AcademicSpaceID, assignment.ProfessorID,
		assignment.RoleInAssignment, assignment.ContractedHoursPerWeek,
	)
	if err := row.Scan(&assignment.ID, &assignment.CreatedAt, &assignment.UpdatedAt); err != nil {
		return fmt.Errorf("assignment Create: %w", err)
	}
	return nil
}

func (assignmentRepo *AssignmentRepo) FindByID(ctx context.Context, id int64) (*domain.Assignment, error) {
	const query = `
		SELECT id, user_id, academic_space_id, professor_id,
		role_in_assignment, contracted_hours_per_week, created_at, updated_at
		FROM assignments
		WHERE id = $1`

	row := assignmentRepo.pool.inner.QueryRow(ctx, query, id)
	assignment, err := scanAssignment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrVinculacionNoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("assignment FindByID: %w", err)
	}
	return assignment, nil
}

func (assignmentRepo *AssignmentRepo) FindBySpace(ctx context.Context, spaceID int64) ([]domain.Assignment, error) {
	const query = `
		SELECT id, user_id, academic_space_id, professor_id,
		role_in_assignment, contracted_hours_per_week, created_at, updated_at
		FROM assignments
		WHERE academic_space_id = $1
		ORDER BY created_at ASC`

	return assignmentRepo.queryMany(ctx, query, spaceID)
}

func (assignmentRepo *AssignmentRepo) FindByUser(ctx context.Context, userID int64) ([]domain.Assignment, error) {
	const query = `
		SELECT id, user_id, academic_space_id, professor_id,
		role_in_assignment, contracted_hours_per_week, created_at, updated_at
		FROM assignments
		WHERE user_id = $1
		ORDER BY created_at ASC`

	return assignmentRepo.queryMany(ctx, query, userID)
}

func (assignmentRepo *AssignmentRepo) FindActiveByUserAndRole(ctx context.Context, userID int64, role string) ([]domain.Assignment, error) {
	const query = `
		SELECT assignment.id, assignment.user_id, assignment.academic_space_id, assignment.professor_id,
		assignment.role_in_assignment, assignment.contracted_hours_per_week, assignment.created_at, assignment.updated_at
		FROM assignments assignment
		JOIN academic_spaces space ON space.id = assignment.academic_space_id
		WHERE assignment.user_id = $1
		AND assignment.role_in_assignment = $2
		AND space.status = 'active'
		ORDER BY assignment.created_at ASC`

	return assignmentRepo.queryMany(ctx, query, userID, role)
}

func (assignmentRepo *AssignmentRepo) ExistsByUserSpaceRole(ctx context.Context, userID, spaceID int64, role string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM assignments
			WHERE user_id = $1 AND academic_space_id = $2 AND role_in_assignment = $3
		)`

	var exists bool
	if err := assignmentRepo.pool.inner.QueryRow(ctx, query, userID, spaceID, role).Scan(&exists); err != nil {
		return false, fmt.Errorf("assignment ExistsByUserSpaceRole: %w", err)
	}
	return exists, nil
}

func (assignmentRepo *AssignmentRepo) queryMany(ctx context.Context, query string, args ...any) ([]domain.Assignment, error) {
	rows, err := assignmentRepo.pool.inner.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("assignment query: %w", err)
	}
	defer rows.Close()

	var assignments []domain.Assignment
	for rows.Next() {
		assignment, err := scanAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("assignment scan: %w", err)
		}
		assignments = append(assignments, *assignment)
	}
	return assignments, rows.Err()
}

func scanAssignment(row scannable) (*domain.Assignment, error) {
	assignment := &domain.Assignment{}
	err := row.Scan(
		&assignment.ID, &assignment.UserID, &assignment.AcademicSpaceID, &assignment.ProfessorID,
		&assignment.RoleInAssignment, &assignment.ContractedHoursPerWeek,
		&assignment.CreatedAt, &assignment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return assignment, nil
}

func (assignmentRepo *AssignmentRepo) FindByProfessorWithUser(ctx context.Context, professorID int64) ([]domain.AssignmentWithUser, error) {
	const query = `
		SELECT a.id, a.user_id, a.academic_space_id, a.professor_id,
		a.role_in_assignment, a.contracted_hours_per_week,
		a.created_at, a.updated_at,
		u.name, u.email
		FROM assignments a
		JOIN users u ON u.id = a.user_id
		WHERE a.professor_id = $1
		ORDER BY a.created_at ASC`

	rows, err := assignmentRepo.pool.inner.Query(ctx, query, professorID)
	if err != nil {
		return nil, fmt.Errorf("assignment FindByProfessorWithUser: %w", err)
	}
	defer rows.Close()

	var results []domain.AssignmentWithUser
	for rows.Next() {
		var item domain.AssignmentWithUser
		err := rows.Scan(
			&item.ID, &item.UserID, &item.AcademicSpaceID, &item.ProfessorID,
			&item.RoleInAssignment, &item.ContractedHoursPerWeek,
			&item.CreatedAt, &item.UpdatedAt,
			&item.UserName, &item.UserEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("assignment scan: %w", err)
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (assignmentRepo *AssignmentRepo) Update(ctx context.Context, assignment *domain.Assignment) error {
	const query = `
		UPDATE assignments
		SET role_in_assignment = $1, contracted_hours_per_week = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, created_at, updated_at`

	row := assignmentRepo.pool.inner.QueryRow(ctx, query,
		assignment.RoleInAssignment, assignment.ContractedHoursPerWeek, assignment.ID,
	)
	if err := row.Scan(&assignment.ID, &assignment.CreatedAt, &assignment.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVinculacionNoEncontrada
		}
		return fmt.Errorf("assignment Update: %w", err)
	}
	return nil
}
