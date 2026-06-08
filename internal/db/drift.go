package db

import (
	"context"
	"time"

	"github.com/Alexmaster12345/infrabios/internal/models"
)

func (d *DB) CreateDriftEvent(ctx context.Context, id string, ev *models.DriftEvent) error {
	_, err := d.pool.Exec(ctx,
		`INSERT INTO drift_events (id, server_id, setting_key, expected_value, actual_value)
		 VALUES ($1,$2,$3,$4,$5)`,
		id, ev.ServerID, ev.SettingKey, []byte(ev.ExpectedValue), []byte(ev.ActualValue))
	return err
}

func (d *DB) ListDriftEvents(ctx context.Context, serverID string, status string) ([]*models.DriftEvent, error) {
	q := `SELECT id, server_id, setting_key, expected_value, actual_value, detected_at, resolved_at, status
	      FROM drift_events WHERE 1=1`
	args := []any{}
	i := 1
	if serverID != "" {
		q += ` AND server_id=$` + itoa(i)
		args = append(args, serverID)
		i++
	}
	if status != "" {
		q += ` AND status=$` + itoa(i)
		args = append(args, status)
	}
	q += ` ORDER BY detected_at DESC`
	rows, err := d.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.DriftEvent
	for rows.Next() {
		var e models.DriftEvent
		var resolvedAt *time.Time
		if err := rows.Scan(&e.ID, &e.ServerID, &e.SettingKey,
			&e.ExpectedValue, &e.ActualValue, &e.DetectedAt, &resolvedAt, &e.Status); err != nil {
			return nil, err
		}
		e.ResolvedAt = resolvedAt
		out = append(out, &e)
	}
	return out, rows.Err()
}

func (d *DB) ResolveDriftEvent(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE drift_events SET status='resolved', resolved_at=NOW() WHERE id=$1`, id)
	return err
}

func (d *DB) IgnoreDriftEvent(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE drift_events SET status='ignored' WHERE id=$1`, id)
	return err
}

func itoa(i int) string {
	return string(rune('0' + i))
}
