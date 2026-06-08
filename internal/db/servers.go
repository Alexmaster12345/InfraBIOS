package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) CreateServer(ctx context.Context, id string, req *models.CreateServerRequest) (*models.Server, error) {
	const q = `
		INSERT INTO servers (id, hostname, ip_address, vendor, model, serial_number, bios_version, bmc_version, device_type, os)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, hostname, ip_address, vendor, model, serial_number, bios_version, bmc_version,
		          device_type, os, profile_id, status, last_seen, created_at, updated_at`
	row := d.pool.QueryRow(ctx, q,
		id, req.Hostname, req.IPAddress, req.Vendor, req.Model,
		req.SerialNumber, req.BIOSVersion, req.BMCVersion, req.DeviceType, req.OS)
	return scanServer(row)
}

func (d *DB) GetServer(ctx context.Context, id string) (*models.Server, error) {
	const q = `SELECT id, hostname, ip_address, vendor, model, serial_number, bios_version, bmc_version,
	                  device_type, os, profile_id, status, last_seen, created_at, updated_at
	           FROM servers WHERE id = $1`
	return scanServer(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) GetServerByHostname(ctx context.Context, hostname string) (*models.Server, error) {
	const q = `SELECT id, hostname, ip_address, vendor, model, serial_number, bios_version, bmc_version,
	                  device_type, os, profile_id, status, last_seen, created_at, updated_at
	           FROM servers WHERE hostname = $1`
	return scanServer(d.pool.QueryRow(ctx, q, hostname))
}

func (d *DB) ListServers(ctx context.Context) ([]*models.Server, error) {
	const q = `SELECT id, hostname, ip_address, vendor, model, serial_number, bios_version, bmc_version,
	                  device_type, os, profile_id, status, last_seen, created_at, updated_at
	           FROM servers ORDER BY hostname`
	rows, err := d.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Server
	for rows.Next() {
		s, err := scanServer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) AssignProfile(ctx context.Context, serverID, profileID string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE servers SET profile_id=$1, updated_at=NOW() WHERE id=$2`,
		profileID, serverID)
	return err
}

func (d *DB) UpdateServerSeen(ctx context.Context, serverID string, biosVersion, bmcVersion string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE servers SET last_seen=NOW(), bios_version=$1, bmc_version=$2, updated_at=NOW() WHERE id=$3`,
		biosVersion, bmcVersion, serverID)
	return err
}

func (d *DB) DeleteServer(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, `DELETE FROM servers WHERE id=$1`, id)
	return err
}

// BIOSSettings

func (d *DB) SaveSettings(ctx context.Context, id, serverID string, req *models.SubmitInventoryRequest) (*models.BIOSSettings, error) {
	const q = `INSERT INTO bios_settings (id, server_id, settings, agent_version)
	           VALUES ($1,$2,$3,$4)
	           RETURNING id, server_id, settings, agent_version, collected_at`
	row := d.pool.QueryRow(ctx, q, id, serverID, []byte(req.Settings), req.AgentVersion)
	var s models.BIOSSettings
	err := row.Scan(&s.ID, &s.ServerID, &s.Settings, &s.AgentVersion, &s.CollectedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) LatestSettings(ctx context.Context, serverID string) (*models.BIOSSettings, error) {
	const q = `SELECT id, server_id, settings, agent_version, collected_at
	           FROM bios_settings WHERE server_id=$1 ORDER BY collected_at DESC LIMIT 1`
	var s models.BIOSSettings
	err := d.pool.QueryRow(ctx, q, serverID).
		Scan(&s.ID, &s.ServerID, &s.Settings, &s.AgentVersion, &s.CollectedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) ListSettings(ctx context.Context, serverID string, limit int) ([]*models.BIOSSettings, error) {
	const q = `SELECT id, server_id, settings, agent_version, collected_at
	           FROM bios_settings WHERE server_id=$1 ORDER BY collected_at DESC LIMIT $2`
	rows, err := d.pool.Query(ctx, q, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.BIOSSettings
	for rows.Next() {
		var s models.BIOSSettings
		if err := rows.Scan(&s.ID, &s.ServerID, &s.Settings, &s.AgentVersion, &s.CollectedAt); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

// helpers

type scanner interface {
	Scan(dest ...any) error
}

func scanServer(row scanner) (*models.Server, error) {
	var s models.Server
	var lastSeen *time.Time
	err := row.Scan(
		&s.ID, &s.Hostname, &s.IPAddress, &s.Vendor, &s.Model,
		&s.SerialNumber, &s.BIOSVersion, &s.BMCVersion,
		&s.DeviceType, &s.OS,
		&s.ProfileID, &s.Status, &lastSeen, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.LastSeen = lastSeen
	return &s, nil
}
