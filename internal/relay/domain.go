package relay

import (
	"encoding/json"
	"time"
)

type OutBoxEvents struct {
	ID      int64
	Payload json.RawMessage
	EventId string
	Status string
	CratedAt time.Time
}

type Message struct {
	OutboxID int64
	ID      string
	Payload json.RawMessage
	EventType string
	Status string
}