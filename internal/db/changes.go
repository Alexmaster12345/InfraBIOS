package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) CreateChange(ctx context.Context, id string, req *models.CreateChangeRequest) (*models.ChangeRequest, error) {
	const q = `INSERT INTO change_requests (id, server_id, profile_id, requested_by, type, payload)
	           VALUES ($1,$2,$3,$4,$5,$6)
	           RETURNING id, server_id, profile_id, requested_by, type, payload,
	                     status, requested_at, reviewed_by, reviewed_at, applied_at`
	row := d.pool.QueryRow(ctx, q, id, req.ServerID, req.ProfileID, req.RequestedBy, req.Type, []byte(req.Payload))
	return scanChange(row)
}

func (d *DB) GetChange(ctx context.Context, id string) (*models.ChangeRequest, error) {
	const q = `SELECT id, server_id, profile_id, requested_by, type, payload,
	                  status, requested_at, reviewed_by, reviewed_at, applied_at
	           FROM change_requests WHERE id=$1`
	return scanChange(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) ListChanges(ctx context.Context, status string) ([]*models.ChangeRequest, error) {
	q := `SELECT id, server_id, profile_id, requested_by, type, payload,
	             status, requested_at, reviewed_by, reviewed_at, applied_at
	      FROM change_requests`
	var args []any
	if status != "" {
		q += ` WHERE status=$1`
		args = append(args, status)
	}
	q += ` ORDER BY requested_at DESC`
	rows, err := d.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.ChangeRequest
	for rows.Next() {
		c, err := scanChange(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *DB) ReviewChange(ctx context.Context, id string, req *models.ReviewChangeRequest) error {
	status := "rejected"
	if req.Approved {
		status = "approved"
	}
	_, err := d.pool.Exec(ctx,
		`UPDATE change_requests SET status=$1, reviewed_by=$2, reviewed_at=NOW() WHERE id=$3`,
		status, req.ReviewedBy, id)
	return err
}

func (d *DB) MarkChangeApplied(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE change_requests SET status='applied', applied_at=NOW() WHERE id=$1`, id)
	return err
}

func (d *DB) MarkChangeFailed(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE change_requests SET status='failed' WHERE id=$1`, id)
	return err
}

func scanChange(row scanner) (*models.ChangeRequest, error) {
	var c models.ChangeRequest
	var reviewedAt *time.Time
	var appliedAt *time.Time
	err := row.Scan(
		&c.ID, &c.ServerID, &c.ProfileID, &c.RequestedBy, &c.Type, &c.Payload,
		&c.Status, &c.RequestedAt, &c.ReviewedBy, &reviewedAt, &appliedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.ReviewedAt = reviewedAt
	c.AppliedAt = appliedAt
	return &c, nil
}
