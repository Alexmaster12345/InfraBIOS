package db

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) CreateSnapshot(ctx context.Context, id, serverID string, req *models.CreateSnapshotRequest, settings, firmware []byte) (*models.Snapshot, error) {
	const q = `INSERT INTO snapshots (id, server_id, settings, firmware, taken_by, description)
	           VALUES ($1,$2,$3,$4,$5,$6)
	           RETURNING id, server_id, settings, firmware, taken_by, taken_at, description`
	row := d.pool.QueryRow(ctx, q, id, serverID, settings, firmware, req.TakenBy, req.Description)
	return scanSnapshot(row)
}

func (d *DB) GetSnapshot(ctx context.Context, id string) (*models.Snapshot, error) {
	const q = `SELECT id, server_id, settings, firmware, taken_by, taken_at, description
	           FROM snapshots WHERE id=$1`
	return scanSnapshot(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) ListSnapshots(ctx context.Context, serverID string) ([]*models.Snapshot, error) {
	const q = `SELECT id, server_id, settings, firmware, taken_by, taken_at, description
	           FROM snapshots WHERE server_id=$1 ORDER BY taken_at DESC`
	rows, err := d.pool.Query(ctx, q, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Snapshot
	for rows.Next() {
		s, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) DeleteSnapshot(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, `DELETE FROM snapshots WHERE id=$1`, id)
	return err
}

func scanSnapshot(row scanner) (*models.Snapshot, error) {
	var s models.Snapshot
	err := row.Scan(&s.ID, &s.ServerID, &s.Settings, &s.Firmware, &s.TakenBy, &s.TakenAt, &s.Description)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
