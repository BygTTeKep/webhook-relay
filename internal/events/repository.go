package events

import (
	"context"
	"database/sql"
	"fmt"
	"webhook-relay/internal/database"

	"github.com/google/uuid"
)

type Repository struct {
	*database.PGRepository
}

func NewRepository(pg *database.PGRepository) *Repository {
	return &Repository{
		pg,
	}
}

func (er *Repository) Publish(ctx context.Context, e Event) error {
 	return er.WithTx(ctx, func(tx *sql.Tx) error {
		if err := er.SaveEvent(ctx, tx, e); err != nil {
			return err
		}
		return er.SaveOutbox(ctx, tx, e)
	})
}

func (er *Repository) SaveEvent(ctx context.Context, tx *sql.Tx, e Event) error {
	query := `
		INSERT INTO events(id, event_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(ctx, query,uuid.NewString(), e.EventID, e.EventType, e.Payload, e.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert to events: %s", err)
	}
	return  nil
}

func (er *Repository) SaveOutbox(ctx context.Context, tx *sql.Tx, e Event) error {
	query := `
		INSERT INTO outbox(event_id, status, created_at)
		VALUES ($1, 'pending', $2)
	`
	_, err := tx.ExecContext(ctx, query, e.EventID, e.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return  nil
}