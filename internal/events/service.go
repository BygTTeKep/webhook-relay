package events

import (
	"context"
	"database/sql"
)

type RepositoryInterface interface {
	SaveEvent(ctx context.Context, tx *sql.Tx, e Event) error
	Publish(ctx context.Context, e Event) error
	SaveOutbox(ctx context.Context, tx *sql.Tx, e Event) error
}


type Service struct {
	repo RepositoryInterface
}

func NewEventService(repo RepositoryInterface) *Service {
	return &Service{
		repo: repo,
	}
}

func (es *Service) Publish(ctx context.Context, dto CreateEventRequestDto) (Event, error){
	event := NewEvent(dto.EventType, dto.Payload)
	err := es.repo.Publish(ctx, event)
	if err != nil {
		return Event{}, err
	}
	return event, nil
}