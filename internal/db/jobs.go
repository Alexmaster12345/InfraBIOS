package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) CreateJob(ctx context.Context, id string, req *models.CreateJobRequest) (*models.Job, error) {
	const q = `INSERT INTO jobs (id, type, target, created_by)
	           VALUES ($1,$2,$3,$4)
	           RETURNING id, type, target, status, progress, result, created_by, created_at, started_at, completed_at`
	row := d.pool.QueryRow(ctx, q, id, req.Type, []byte(req.Target), req.CreatedBy)
	return scanJob(row)
}

func (d *DB) GetJob(ctx context.Context, id string) (*models.Job, error) {
	const q = `SELECT id, type, target, status, progress, result, created_by, created_at, started_at, completed_at
	           FROM jobs WHERE id=$1`
	return scanJob(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) ListJobs(ctx context.Context, status string) ([]*models.Job, error) {
	q := `SELECT id, type, target, status, progress, result, created_by, created_at, started_at, completed_at
	      FROM jobs`
	var args []any
	if status != "" {
		q += ` WHERE status=$1`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC LIMIT 100`
	rows, err := d.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (d *DB) StartJob(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET status='running', started_at=NOW() WHERE id=$1`, id)
	return err
}

func (d *DB) UpdateJobProgress(ctx context.Context, id string, progress int) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET progress=$1 WHERE id=$2`, progress, id)
	return err
}

func (d *DB) CompleteJob(ctx context.Context, id string, result []byte) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET status='completed', progress=100, result=$1, completed_at=NOW() WHERE id=$2`,
		result, id)
	return err
}

func (d *DB) FailJob(ctx context.Context, id string, result []byte) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET status='failed', result=$1, completed_at=NOW() WHERE id=$2`,
		result, id)
	return err
}

func (d *DB) CancelJob(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET status='cancelled', completed_at=NOW() WHERE id=$1 AND status IN ('queued','running')`, id)
	return err
}

func scanJob(row scanner) (*models.Job, error) {
	var j models.Job
	var startedAt, completedAt *time.Time
	err := row.Scan(
		&j.ID, &j.Type, &j.Target, &j.Status, &j.Progress, &j.Result,
		&j.CreatedBy, &j.CreatedAt, &startedAt, &completedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	j.StartedAt = startedAt
	j.CompletedAt = completedAt
	return &j, nil
}
