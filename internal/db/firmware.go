package db

import (
	"context"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) UpsertFirmware(ctx context.Context, id, serverID string, req *models.UpsertFirmwareRequest) (*models.FirmwareEntry, error) {
	const q = `INSERT INTO firmware_inventory (id, server_id, component, current_version, approved_version, status)
	           VALUES ($1,$2,$3,$4,$5,
	               CASE WHEN $4 = $5 THEN 'approved' ELSE 'outdated' END
	           )
	           ON CONFLICT (server_id, component) DO UPDATE
	               SET current_version=$4,
	                   approved_version=CASE WHEN firmware_inventory.approved_version='' THEN $5 ELSE firmware_inventory.approved_version END,
	                   status=CASE WHEN $4 = firmware_inventory.approved_version THEN 'approved' ELSE 'outdated' END,
	                   updated_at=NOW()
	           RETURNING id, server_id, component, current_version, approved_version, status, updated_at`
	var f models.FirmwareEntry
	err := d.pool.QueryRow(ctx, q, id, serverID, req.Component, req.CurrentVersion, req.ApprovedVersion).
		Scan(&f.ID, &f.ServerID, &f.Component, &f.CurrentVersion, &f.ApprovedVersion, &f.Status, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (d *DB) ListFirmware(ctx context.Context, serverID string) ([]*models.FirmwareEntry, error) {
	const q = `SELECT id, server_id, component, current_version, approved_version, status, updated_at
	           FROM firmware_inventory WHERE server_id=$1 ORDER BY component`
	rows, err := d.pool.Query(ctx, q, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.FirmwareEntry
	for rows.Next() {
		var f models.FirmwareEntry
		if err := rows.Scan(&f.ID, &f.ServerID, &f.Component, &f.CurrentVersion, &f.ApprovedVersion, &f.Status, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &f)
	}
	return out, rows.Err()
}

func (d *DB) ApproveFirmware(ctx context.Context, serverID, component, version string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE firmware_inventory SET approved_version=$1,
		    status=CASE WHEN current_version=$1 THEN 'approved' ELSE 'outdated' END,
		    updated_at=NOW()
		 WHERE server_id=$2 AND component=$3`,
		version, serverID, component)
	return err
}
