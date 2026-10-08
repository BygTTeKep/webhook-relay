package subs

import (
	"encoding/json"
	"net/http"
	"webhook-relay/internal/httpx"

	"go.uber.org/zap"
)

type Handler struct {
	Service *Service
	Logger  *zap.Logger
}

func NewHandler(service *Service, l *zap.Logger) *Handler {
	return &Handler{
		Service: service,
		Logger:  l,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscription", httpx.Wrap(h.Logger, h.create))
	mux.HandleFunc("GET /subscription/", httpx.Wrap(h.Logger, h.get))
	mux.HandleFunc("POST /subscription/webhook", httpx.Wrap(h.Logger, h.webhookTest))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var req CreateSubscriptionRequestDto
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.BadRequest("invalid json", err)
	}
	if err := req.Validate(); err != nil {
		return httpx.BadRequest(err.Error(), err)
	}
	err := h.Service.Create(ctx, req)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w)
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	var req GetSubscriptionRequestDto
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.BadRequest("Invalid json", err)
	}
	sub, err := h.Service.Get(ctx, req)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sub)
	return nil
}

func (h *Handler) webhookTest(w http.ResponseWriter, r *http.Request) error {
	var j json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		return httpx.BadRequest("Invalid json", err)
	}
	h.Logger.Info("req", zap.Any("body", j))
	json.NewEncoder(w).Encode("success")
	return nil
}
