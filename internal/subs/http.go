package subs

import (
	"errors"
	"fmt"
	"slices"
	"webhook-relay/internal/events"
)

type CreateSubscriptionRequestDto struct {
	Events []string `json:"events"`
	URL    string   `json:"url"`
}

type GetSubscriptionRequestDto struct {
	ID string `json:"id"`
	Secret string `json:"secret"`
}


func (csr *CreateSubscriptionRequestDto) Validate() error {
	var errs []error
	if (len(csr.URL) == 0) {
		errs = append(errs, errors.New("url must not be empty"))
	}
	if (len(csr.Events) == 0) {
		errs = append(errs, errors.New("events must not be empty"))
	}
	for _, event := range csr.Events {
		if !slices.Contains(events.AllowedEvents, event) {
			errs = append(errs, fmt.Errorf("unknown events: %s", event))
		}
	}
	return errors.Join(errs...)
}