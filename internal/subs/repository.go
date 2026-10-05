package subs

import (
	"context"
	"database/sql"
	"webhook-relay/internal/events"
)

type Repository struct{
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}


func (sr *Repository) Save(ctx context.Context, s Subscription)error {
	querySub := `
	INSERT INTO subscriptions (URL, Secret, Active, CreatedAt, ID)
	VALUES ()`
	_, err:= sr.DB.ExecContext(ctx, querySub, s.URL, s.Secret, s.Active, s.ID)
	if err != nil {
		return  err
	}
	return nil
}

func (sr *Repository) Get(ctx context.Context, id,  secret string) (Subscription, error) {
	query := `
		SELECT * FROM subscriptions
		WHERE id = $1 AND secret = $2
	`
	row := sr.DB.QueryRowContext(ctx, query, id, secret)
	var subscription Subscription
	if err := row.Scan(subscription); err !=nil{
		return Subscription{}, err
	}
	return  subscription, nil
}

func (sr *Repository) FindByEventType(ctx context.Context, t events.EventType) ([]Subscription, error) {
	var subs []Subscription
	query := `
		SELECT FROM subscriptions WHERE 
	`
	rows, err := sr.DB.QueryContext(ctx, query, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err != nil {}
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(&sub); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}