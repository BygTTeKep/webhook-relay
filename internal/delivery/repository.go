package delivery

import (
	"context"
	"fmt"
	"webhook-relay/internal/database"
)

type Repository struct {
	*database.PGRepository
}

func NewRepository(db *database.PGRepository) *Repository{
	return &Repository{db}
}

func (r *Repository) Save(ctx context.Context, d Delivery) error {
	query := `
		INSERT INTO delivery(event_id, subscription_id, status, attempts, last_attempt_at)
		VALUES($1, $2, $3, $4, $5)
	`
	res, err := r.DB.ExecContext(ctx, query, d.EventID, d.SubscriptionID, d.Status, d.Attempts, d.LastAttemptAt)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("delivery save: ", err)
	}
	return nil
}