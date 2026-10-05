package delivery

import "time"

type Delivery struct {
	ID             string
	EventID        string
	SubscriptionID string
	Status         string
	Attempts       int
	LastAttemptAt  time.Time
}