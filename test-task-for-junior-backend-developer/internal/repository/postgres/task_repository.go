package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, recurrence, source_task_id, scheduled_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (source_task_id, scheduled_for)
		DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id, title, description, status, recurrence, source_task_id, scheduled_for, created_at, updated_at
	`

	recurrenceJSON, err := marshalRecurrence(task.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, recurrenceJSON, task.SourceTaskID, task.ScheduledFor, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence, source_task_id, scheduled_for, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			recurrence = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, recurrence, source_task_id, scheduled_for, created_at, updated_at
	`

	recurrenceJSON, err := marshalRecurrence(task.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, recurrenceJSON, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence, source_task_id, scheduled_for, created_at, updated_at
		FROM tasks
		WHERE source_task_id IS NULL
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence, source_task_id, scheduled_for, created_at, updated_at
		FROM tasks
		WHERE recurrence IS NOT NULL AND source_task_id IS NULL
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status         string
		recurrenceJSON []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceJSON,
		&task.SourceTaskID,
		&task.ScheduledFor,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if len(recurrenceJSON) > 0 {
		var recurrence taskdomain.Recurrence
		if err := json.Unmarshal(recurrenceJSON, &recurrence); err != nil {
			return nil, err
		}
		task.Recurrence = &recurrence
	}

	return &task, nil
}

func marshalRecurrence(recurrence *taskdomain.Recurrence) ([]byte, error) {
	if recurrence == nil {
		return nil, nil
	}
	return json.Marshal(recurrence)
}
