package db

import (
	"context"
	"encoding/json"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) SaveComplianceReport(ctx context.Context, id string, r *models.ComplianceReport) error {
	v, err := json.Marshal(r.Violations)
	if err != nil {
		return err
	}
	_, err = d.pool.Exec(ctx,
		`INSERT INTO compliance_reports (id, server_id, profile_id, compliant, violations, score)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		id, r.ServerID, r.ProfileID, r.Compliant, v, r.Score)
	return err
}

func (d *DB) LatestComplianceReport(ctx context.Context, serverID string) (*models.ComplianceReport, error) {
	const q = `SELECT id, server_id, profile_id, compliant, violations, score, generated_at
	           FROM compliance_reports WHERE server_id=$1 ORDER BY generated_at DESC LIMIT 1`
	var r models.ComplianceReport
	var violations []byte
	err := d.pool.QueryRow(ctx, q, serverID).
		Scan(&r.ID, &r.ServerID, &r.ProfileID, &r.Compliant, &violations, &r.Score, &r.GeneratedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(violations, &r.Violations); err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) ListComplianceReports(ctx context.Context, serverID string) ([]*models.ComplianceReport, error) {
	const q = `SELECT id, server_id, profile_id, compliant, violations, score, generated_at
	           FROM compliance_reports WHERE server_id=$1 ORDER BY generated_at DESC LIMIT 50`
	rows, err := d.pool.Query(ctx, q, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.ComplianceReport
	for rows.Next() {
		var r models.ComplianceReport
		var violations []byte
		if err := rows.Scan(&r.ID, &r.ServerID, &r.ProfileID, &r.Compliant, &violations, &r.Score, &r.GeneratedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(violations, &r.Violations)
		out = append(out, &r)
	}
	return out, rows.Err()
}
