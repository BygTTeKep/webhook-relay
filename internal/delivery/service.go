package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"webhook-relay/internal/events"
	"webhook-relay/internal/subs"
)

type RepositoryInterface interface {
	Save(ctx context.Context, d Delivery) error
}

type Service struct {
	repo RepositoryInterface
	client *http.Client
}

func NewService(repo RepositoryInterface, client *http.Client) *Service {
	return &Service{
		repo: repo,
		client: client,
	}
}

func (s *Service) Delivery(ctx context.Context, e events.Event, sub subs.Subscription) error {
	body, _ := json.Marshal(e)
	sig := sign(body, sub.Secret)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL, bytes.NewReader(body))
	req.Header.Set("X-Signature", sig)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	status := "success"
	if err != nil || resp.StatusCode >= 400 {
		status = "failed"
	}
	err = s.repo.Save(ctx, Delivery{
		EventID:        e.ID,
		SubscriptionID: sub.ID,
		Status:         status,
		Attempts:       1,
		LastAttemptAt:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("delivery event save failed")
	}
	return nil
}