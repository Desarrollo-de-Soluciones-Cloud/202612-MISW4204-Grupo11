package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *taskRepository {
	return &taskRepository{db: db}
}

func (repository *taskRepository) Create(task *domain.Task) error {
	const query = `
		INSERT INTO tasks (
			title,
			description,
			status,
			week,
			time_invested,
			assignment_id,
			time_registered,
			observations
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`

	err := repository.db.QueryRow(
		context.Background(),
		query,
		task.Title,
		task.Description,
		string(task.Status),
		task.Week,
		task.TimeInvested,
		task.Assignment_id,
		task.TimeRegistered,
		task.Observations,
	).Scan(&task.ID)

	if err != nil {
		return fmt.Errorf("error creating task: %w", err)
	}

	return nil
}

func (repository *taskRepository) GetAll() ([]domain.Task, error) {
	const query = `
		SELECT
			id,
			title,
			description,
			status,
			week,
			time_invested,
			time_registered,
			observations
		FROM tasks
		ORDER BY id;
	`

	rows, err := repository.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks: %w", err)
	}
	defer rows.Close()

	tasks := []domain.Task{}

	for rows.Next() {
		var task domain.Task
		var status string

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&status,
			&task.Week,
			&task.TimeInvested,
			&task.TimeRegistered,
			&task.Observations,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning task: %w", err)
		}

		task.Status = domain.Status(status)
		tasks = append(tasks, task)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", rows.Err())
	}

	return tasks, nil
}

func (repository *taskRepository) GetByID(id string) (*domain.Task, error) {
	taskID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}

	const query = `
		SELECT
			id,
			title,
			description,
			status,
			week,
			time_invested,
			time_registered,
			observations
		FROM tasks
		WHERE id = $1;
	`

	var task domain.Task
	var status string

	err = repository.db.QueryRow(context.Background(), query, taskID).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.Week,
		&task.TimeInvested,
		&task.TimeRegistered,
		&task.Observations,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("task with id %d not found", taskID)
		}
		return nil, fmt.Errorf("error getting task by id: %w", err)
	}

	task.Status = domain.Status(status)

	return &task, nil
}

func (repository *taskRepository) Update(task *domain.Task) error {
	const query = `
		UPDATE tasks
		SET
			title = $1,
			description = $2,
			status = $3,
			week = $4,
			time_invested = $5,
			time_registered = $6,
			observations = $7
		WHERE id = $8;
	`

	commandTag, err := repository.db.Exec(
		context.Background(),
		query,
		task.Title,
		task.Description,
		string(task.Status),
		task.Week,
		task.TimeInvested,
		task.TimeRegistered,
		task.Observations,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating task: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("task with id %d not found", task.ID)
	}

	return nil
}

func (repository *taskRepository) Delete(id string) error {
	taskID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	const query = `
		DELETE FROM tasks
		WHERE id = $1;
	`

	commandTag, err := repository.db.Exec(context.Background(), query, taskID)
	if err != nil {
		return fmt.Errorf("error deleting task: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("task with id %d not found", taskID)
	}

	return nil
}

func (repository *taskRepository) UpdateStatus(task *domain.Task) error {
	const query = `
		UPDATE tasks
		SET status = $1
		WHERE id = $2;
	`

	commandTag, err := repository.db.Exec(
		context.Background(),
		query,
		string(task.Status),
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating task status: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("task with id %d not found", task.ID)
	}

	return nil
}

func (repository *taskRepository) SaveAttachment(attachment *domain.Attachment) error {
	const query = `
		INSERT INTO attachments (
			task_id,
			file_name,
			content_type,
			storage_path
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := repository.db.QueryRow(
		context.Background(),
		query,
		attachment.TaskID,
		attachment.FileName,
		attachment.ContentType,
		attachment.StoragePath,
	).Scan(&attachment.ID)
	if err != nil {
		return fmt.Errorf("error saving attachment: %w", err)
	}

	return nil
}
