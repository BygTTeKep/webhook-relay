package subs

import (
	"context"
	"fmt"
	"time"
	"webhook-relay/internal/database"
	"webhook-relay/internal/events"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

type Repository struct{
	*database.PGRepository
}

func NewRepo(db *database.PGRepository) *Repository {
	return &Repository{
		db,
	}
}


func (sr *Repository) Save(ctx context.Context, s Subscription, tx pgx.Tx) (string, error) {
	query := `
		INSERT INTO subscriptions (url, secret, active, created_at, id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var insertId string
	uuid := uuid.NewString()
	err := tx.QueryRow(ctx, query, s.URL, s.Secret, s.Active, time.Now(), uuid).Scan(&insertId)
	if err != nil {
		return insertId, err
	}
	return insertId, nil
}

func (sr *Repository) SaveSubEventsTx(ctx context.Context, subId string, e []string, tx pgx.Tx) error {
	if len(e) == 0 {
		return nil
	}

	query := `
		INSERT INTO subscription_events(subscription_id, event_type)
		SELECT $1, unnest($2::text[])
	`
	if _, err := tx.Exec(ctx, query, subId, pq.Array(e)); err != nil {
		return fmt.Errorf("insert subscription_events: %w", err)
	}
	return nil
}

func (sr *Repository) SaveSubAndEventsTx(ctx context.Context, s Subscription, e []string) error {
	err := sr.WithTx(ctx, func(tx pgx.Tx) error {
		subId, err := sr.Save(ctx, s, tx)
		if err != nil {
			return err
		}
		err = sr.SaveSubEventsTx(ctx, subId, e, tx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (sr *Repository) Get(ctx context.Context, id,  secret string) (Subscription, error) {
	query := `
		SELECT * FROM subscriptions
		WHERE id = $1 AND secret = $2
	`
	row := sr.DB.QueryRow(ctx, query, id, secret)
	var subscription Subscription
	if err := row.Scan(subscription); err !=nil{
		return Subscription{}, err
	}
	return  subscription, nil
}

func (sr *Repository) FindByEventType(ctx context.Context, t events.EventType) ([]Subscription, error) {
	var subs []Subscription
	query := `
		SELECT
			s.id,
			s.url,
			s.secret,
			s.active,
			s.created_at 
		FROM subscription_events se 
		INNER JOIN subscriptions s on s.id = se.subscription_id
		WHERE se.event_type = $1 AND s.active = true 
	`
	rows, err := sr.DB.Query(ctx, query, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err != nil {}
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(&sub.ID, &sub.URL, &sub.Secret, &sub.Active, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}