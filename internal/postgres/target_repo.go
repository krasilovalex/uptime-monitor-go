package postgres

import (
	"context"
	"uptime-monitor/internal/manager"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TargetRepo struct {
	db *pgxpool.Pool
}

func NewTargetRepository(db *pgxpool.Pool) *TargetRepo {
	return &TargetRepo{db: db}
}

func (r *TargetRepo) GetActiveTargets(ctx context.Context) ([]manager.Target, error) {
	query := `SELECT id, url, interval_seconds, timeout_seconds, is_active, created_at FROM targets WHERE is_active = true`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var targets []manager.Target
	for rows.Next() {
		var t manager.Target
		if err := rows.Scan(&t.ID, &t.URL, &t.IntervalSeconds, &t.TimeoutSeconds, &t.IsActive, &t.CreatedAt); err != nil {
			return nil, err
		}

		targets = append(targets, t)
	}

	return targets, nil
}

func (r *TargetRepo) CreateTarget(ctx context.Context, url string, interval, timeout int) (string, error) {
	query := `INSERT INTO targets (url, inverval_seconds, timeout_seconds)
	VALUES ($1, $2, $3) RETURNING id`

	var id string
	err := r.db.QueryRow(ctx, query, url, interval, timeout).Scan(&id)

	return id, err
}

func (r *TargetRepo) DeleteTarget(ctx context.Context, id string) error {
	query := `DELETE FROM targets WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
