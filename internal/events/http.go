package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

type CreateEventRequestDto struct {
	EventType string              `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type CreateEventResponseDto struct {
	EventID string
}

func (cer *CreateEventRequestDto) Validate() error {
	var errs []error
	if cer.EventType == "" {
		errs = append(errs, fmt.Errorf("event_type is required"))
	} else if !slices.Contains(AllowedEvents, cer.EventType) {
		errs = append(errs, fmt.Errorf("unknown event type: %s", cer.EventType))
	}
	if len(cer.Payload) == 0 {
		errs = append(errs, fmt.Errorf("payload is required"))
	}
	return errors.Join(errs...)
}