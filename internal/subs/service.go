package subs

import (
	"context"
	"fmt"
	"webhook-relay/internal/events"
)

type RepoInterface interface {
	Save(ctx context.Context, s Subscription) error
	Get(ctx context.Context, id, secret string) (Subscription, error)
	FindByEventType(ctx context.Context, t events.EventType) ([]Subscription, error)
}

type Service struct {
	repo RepoInterface
}

func NewService(repo RepoInterface) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, dto CreateSubscriptionRequestDto) error{
	subscription, err := NewSubscription(dto.URL, dto.Events)
	if err != nil {
		return fmt.Errorf("save subscription: %s", err)
	}
	if err := s.repo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("save subscription: %s", err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, dto GetSubscriptionRequestDto) (Subscription ,error) {
	subscription, err := s.repo.Get(ctx, dto.ID, dto.Secret)
	if err != nil {
		return Subscription{}, fmt.Errorf("get subscription by id: %s", err)
	}
	return subscription, nil
}