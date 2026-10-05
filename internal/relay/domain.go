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