package events

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler)Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", h.create)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequestDto
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	event, err := h.service.Publish(ctx, req)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(event)
}