package db

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) CreateProfile(ctx context.Context, id string, req *models.CreateProfileRequest) (*models.BIOSProfile, error) {
	const q = `INSERT INTO bios_profiles (id, name, description, settings, tags, created_by)
	           VALUES ($1,$2,$3,$4,$5,$6)
	           RETURNING id, name, description, settings, tags, created_by, created_at, updated_at`
	row := d.pool.QueryRow(ctx, q, id, req.Name, req.Description, []byte(req.Settings), req.Tags, req.CreatedBy)
	return scanProfile(row)
}

func (d *DB) GetProfile(ctx context.Context, id string) (*models.BIOSProfile, error) {
	const q = `SELECT id, name, description, settings, tags, created_by, created_at, updated_at
	           FROM bios_profiles WHERE id=$1`
	return scanProfile(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) ListProfiles(ctx context.Context) ([]*models.BIOSProfile, error) {
	const q = `SELECT id, name, description, settings, tags, created_by, created_at, updated_at
	           FROM bios_profiles ORDER BY name`
	rows, err := d.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.BIOSProfile
	for rows.Next() {
		p, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.BIOSProfile, error) {
	const q = `UPDATE bios_profiles
	           SET description=$1, settings=$2, tags=$3, updated_at=NOW()
	           WHERE id=$4
	           RETURNING id, name, description, settings, tags, created_by, created_at, updated_at`
	row := d.pool.QueryRow(ctx, q, req.Description, []byte(req.Settings), req.Tags, id)
	return scanProfile(row)
}

func (d *DB) DeleteProfile(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, `DELETE FROM bios_profiles WHERE id=$1`, id)
	return err
}

func scanProfile(row scanner) (*models.BIOSProfile, error) {
	var p models.BIOSProfile
	err := row.Scan(
		&p.ID, &p.Name, &p.Description, &p.Settings, &p.Tags,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
