package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sync"
	"time"
	"webhook-relay/internal/relay"
	"webhook-relay/internal/subs"

	"github.com/sony/gobreaker"
)

const (
	maxAttempts    = 4
	attemptTimeout = 5 * time.Second
)

type RepositoryInterface interface {
	Save(ctx context.Context, d Delivery) error
}

type Service struct {
	repo   RepositoryInterface
	client *http.Client

	mu       sync.Mutex
	breakers map[string]*gobreaker.CircuitBreaker
}

type permanentError struct{ err error }

func (e *permanentError) Error() string { return e.err.Error() }

func NewService(repo RepositoryInterface, client *http.Client) *Service {
	return &Service{
		repo:     repo,
		client:   client,
		breakers: make(map[string]*gobreaker.CircuitBreaker),
	}
}

func (s *Service) Delivery(ctx context.Context, e relay.Message, sub subs.Subscription) error {
	body, err := json.Marshal(e.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	cb := s.breakerFor(sub.URL)
	sig := sign(body, sub.Secret)

	var sendErr error
	attempts := 0
	for attempts < maxAttempts {
		attempts++

		_, sendErr = cb.Execute(func() (interface{}, error) {
			return nil, s.send(ctx, sig, sub, body)
		})
		if sendErr == nil {
			break
		}

		var perm *permanentError
		if errors.As(sendErr, &perm) || errors.Is(sendErr, gobreaker.ErrOpenState) || errors.Is(sendErr, gobreaker.ErrTooManyRequests) {
			break
		}
		if attempts >= maxAttempts {
			break
		}
		t := time.NewTimer(backoff(attempts))
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
	}

	status := "success"
	if sendErr != nil {
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
		return fmt.Errorf("delivery event save failed: %w", err)
	}
	return nil
}

func (s *Service) breakerFor(rawUrl string) *gobreaker.CircuitBreaker {
	key := rawUrl
	if u, err := url.Parse(rawUrl); err == nil {
		key = u.Host
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.breakers[key]; ok {
		return v
	}
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        key,
		MaxRequests: 2,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: func(err error) bool {
			var perErr *permanentError
			return err == nil || errors.As(err, &perErr)
		},
	})
	s.breakers[key] = cb
	return cb
}

func (s *Service) send(ctx context.Context, sig string, sub subs.Subscription, body []byte) error {
	ctx, cancel := context.WithTimeout(ctx, attemptTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL, bytes.NewReader(body))
	if err != nil {
		return &permanentError{err}
	}
	req.Header.Set("X-Signature", sig)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode < 300:
		{
			return nil
		}
	case resp.StatusCode == 408, resp.StatusCode == 429, resp.StatusCode >= 500:
		{
			return fmt.Errorf("subscriber returned %d", resp.StatusCode)
		}
	default:
		return &permanentError{fmt.Errorf("subscriber returned %d", resp.StatusCode)}
	}
}

func backoff(attempt int) time.Duration {
	d := 200 * time.Millisecond << (attempt - 1)
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d/2 + rand.N(d/2)
}
