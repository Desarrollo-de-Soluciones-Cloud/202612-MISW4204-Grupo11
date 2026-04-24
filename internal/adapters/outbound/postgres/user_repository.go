package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (repository *UserRepository) FindCredentialsByEmail(ctx context.Context, email string) (*domain.UserCredentials, error) {
	const findByEmailQuery = `
SELECT u.id, u.name, u.email, u.password_hash, COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}')
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE LOWER(TRIM(u.email)) = LOWER(TRIM($1))
GROUP BY u.id, u.name, u.email, u.password_hash`

	var credentials domain.UserCredentials
	scanErr := repository.db.QueryRow(ctx, findByEmailQuery, email).Scan(
		&credentials.ID, &credentials.Name, &credentials.Email, &credentials.PasswordHash, &credentials.Roles)
	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, scanErr
	}
	return &credentials, nil
}

func (repository *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := repository.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email = $1`, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := repository.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (repository *UserRepository) CreateUser(ctx context.Context, name, email, passwordHash string, roleNames []string) (int64, error) {
	transaction, beginErr := repository.db.Begin(ctx)
	if beginErr != nil {
		return 0, beginErr
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var userID int64
	insertErr := transaction.QueryRow(ctx, `
INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		name, email, passwordHash).Scan(&userID)
	if insertErr != nil {
		return 0, insertErr
	}

	for _, roleName := range roleNames {
		var roleID int64
		lookupErr := transaction.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
		if lookupErr != nil {
			if errors.Is(lookupErr, pgx.ErrNoRows) {
				return 0, fmt.Errorf("unknown role: %s", roleName)
			}
			return 0, lookupErr
		}
		if _, linkErr := transaction.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID); linkErr != nil {
			return 0, linkErr
		}
	}

	if commitErr := transaction.Commit(ctx); commitErr != nil {
		return 0, commitErr
	}
	return userID, nil
}

func (repository *UserRepository) ListUsersByRole(ctx context.Context, role string) ([]domain.User, error) {
	const query = `
SELECT u.id, u.name, u.email,
       COALESCE(array_agg(r.name ORDER BY r.name) FILTER (WHERE r.name IS NOT NULL), '{}')
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.id IN (
    SELECT ur2.user_id FROM user_roles ur2
    JOIN roles r2 ON r2.id = ur2.role_id
    WHERE r2.name = $1
)
GROUP BY u.id
ORDER BY u.id`

	rows, queryErr := repository.db.Query(ctx, query, role)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if scanErr := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Roles); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (repository *UserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	const listUsersQuery = `
SELECT u.id, u.name, u.email,
       COALESCE(array_agg(r.name ORDER BY r.name) FILTER (WHERE r.name IS NOT NULL), '{}')
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
GROUP BY u.id
ORDER BY u.id`

	rows, queryErr := repository.db.Query(ctx, listUsersQuery)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if scanErr := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Roles); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
