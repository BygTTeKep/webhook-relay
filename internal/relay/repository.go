package relay

import (
	"context"
	"database/sql"
	"fmt"
	"webhook-relay/internal/database"

	"github.com/lib/pq"
)

type RepositoryInterface interface {
	FetchPending(ctx context.Context, tx *sql.Tx, limit int)
}

type Repository struct {
	*database.PGRepository
}

func NewRepository(db *database.PGRepository) *Repository {
	return &Repository{db}
}

func (r *Repository) FetchPending(ctx context.Context, tx *sql.Tx, limit int) ([]OutBoxEvents, error) {
	query := `
		SELECT id, event_id, payload, status, created_at
		FROM outbox
		WHERE status = 'pending'
		ORDER BY id
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	row, err := tx.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox: %w", err)
	}
	defer row.Close()
	outbox := make([]OutBoxEvents, 0, limit)

	for row.Next() {
		var o OutBoxEvents
		if err := row.Scan(&o.ID, &o.EventId, &o.Payload, &o.Status, &o.CratedAt); err != nil {
			return nil, fmt.Errorf("scan outbox row: %w", err)
		}
		outbox = append(outbox, o)
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return outbox, nil
}

func (r *Repository) MarkSend(ctx context.Context, tx *sql.Tx, ids []int64) error {
	query := `
		UPDATE outbox SET status='sent' WHERE id = ANY($1)
	`
	_, err := tx.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("mark send: %w", err)
	}
	return  nil
}