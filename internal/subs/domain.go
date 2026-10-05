package subs

import (
	"time"
	"webhook-relay/internal/utils"
)

type Subscription struct {
	ID        string
	URL       string
	Secret    string
	Events     []string
	Active    bool
	CreatedAt time.Time
}

func NewSubscription(URL string, events []string) (Subscription, error) {
	id, err := utils.GenerateId()
	if err != nil {
		return  Subscription{}, err
	}
	secret := utils.GenerateSecret()
	return Subscription{
		URL: URL,
		Events: events,
		Active: true,
		CreatedAt: time.Now(),
		Secret: secret,
		ID: id,
	}, nil
}