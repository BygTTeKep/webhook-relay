package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/go-playground/validator/v10"
)

type CreateEventRequestDto struct {
	EventType string          `json:"event_type" validate:"required"`
	Payload   json.RawMessage `json:"payload" validate:"required,json"`
}

type CreateEventResponseDto struct {
	EventID string
}

func (cer *CreateEventRequestDto) Validate() error {
	var errs []error
	validate := validator.New()

	if err := validate.Struct(cer); err != nil {
		errs = append(errs, err)
	}
	if !slices.Contains(AllowedEvents, cer.EventType) {
		errs = append(errs, fmt.Errorf("unknown event type: %s", cer.EventType))
	}
	return errors.Join(errs...)
}
