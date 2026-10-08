package subs

import (
	"errors"
	"fmt"
	"slices"
	"webhook-relay/internal/events"

	"github.com/go-playground/validator/v10"
)

type CreateSubscriptionRequestDto struct {
	Events []string `json:"events" validate:"required"`
	URL    string   `json:"url" validate:"required,url"`
}

type GetSubscriptionRequestDto struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func (csr *CreateSubscriptionRequestDto) Validate() error {
	var errs []error
	validate := validator.New()
	if err := validate.Struct(csr); err != nil {
		errs = append(errs, err)
	}
	if len(csr.Events) == 0 {
		errs = append(errs, errors.New("events must not be empty"))
	}
	for _, event := range csr.Events {
		if !slices.Contains(events.AllowedEvents, event) {
			errs = append(errs, fmt.Errorf("unknown events: %s", event))
		}
	}
	return errors.Join(errs...)
}
