package events

import (
	"encoding/json"
	"net/http"
	"webhook-relay/internal/httpx"

	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", httpx.Wrap(h.logger, h.create))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var req CreateEventRequestDto
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.BadRequest("invalid json", err)
	}
	if err := req.Validate(); err != nil {
		return httpx.BadRequest(err.Error(), err)
	}
	event, err := h.service.Publish(ctx, req)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(event)
	return nil
}
