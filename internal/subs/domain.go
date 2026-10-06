package subs

import (
	"time"
	"webhook-relay/internal/utils"
)

type Subscription struct {
	ID        string
	URL       string
	Secret    string
	Active    bool
	CreatedAt time.Time
}

type SubscriptionEvents struct {
	ID int
	SubscriptionID string
	EventType string
}

func NewSubscription(URL string, events []string) (Subscription, error) {
	id, err := utils.GenerateId()
	if err != nil {
		return  Subscription{}, err
	}
	secret := utils.GenerateSecret()

	return Subscription{
		URL: URL,
		Active: true,
		CreatedAt: time.Now(),
		Secret: secret,
		ID: id,
	}, nil
}

func NewSubscriptionEvents(sub_id string, events []string) []SubscriptionEvents {
	var subsEvents = make([]SubscriptionEvents, 0, len(events))
	for _, v := range events {
		var subEvent SubscriptionEvents = SubscriptionEvents{
			SubscriptionID: sub_id,
			EventType: v,
		}
		subsEvents = append(subsEvents, subEvent)
	}
	return subsEvents
}