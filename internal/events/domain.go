package events

import (
	"encoding/json"
	"time"
	"webhook-relay/internal/utils"
)

type EventType string

const (
	EventsOrderCreated EventType = "order.created"
	EventsOrderPaid    EventType = "order.paid"
)

var AllowedEvents = []string{"order.created", "order.paid"}

type Event struct {
	ID string
	EventID string
	EventType string
	Payload json.RawMessage
	CreatedAt time.Time
}

func NewEvent(eventType string, payload json.RawMessage) Event {
	eventId := eventType + "_"+ utils.GenerateUUID()
	id := utils.GenerateUUID()
	return Event{
		ID: id,
		EventID: eventId,
		EventType: eventType,
		Payload: payload,
		CreatedAt: time.Now(),
	}
}