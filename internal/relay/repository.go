package relay

import (
	"context"
	"database/sql"
	"fmt"
	"webhook-relay/internal/database"

	"github.com/jackc/pgx/v5"
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

func (r *Repository) FetchPending(ctx context.Context, tx pgx.Tx, limit int) ([]Message, error) {
	query := `
		SELECT 
			o.id as id, 
			e.payload, 
			e.event_type,
			e.id, 
			o.status
		FROM outbox o
		INNER JOIN events e on e.event_id = o.event_id
		WHERE status = 'pending'
		ORDER BY o.id
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	row, err := tx.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox: %w", err)
	}
	defer row.Close()
	outbox := make([]Message, 0, limit)

	for row.Next() {
		var o Message
		if err := row.Scan(&o.OutboxID, &o.Payload, &o.EventType, &o.ID, &o.Status); err != nil {
			return nil, fmt.Errorf("scan outbox row: %w", err)
		}
		outbox = append(outbox, o)
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return outbox, nil
}

func (r *Repository) MarkSend(ctx context.Context, tx pgx.Tx, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := `
		UPDATE outbox SET status='sent' WHERE id = ANY($1)
	`
	_, err := tx.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("mark send: %w", err)
	}
	return  nil
}