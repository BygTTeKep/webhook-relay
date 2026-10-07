package subs

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscription", h.create)
	mux.HandleFunc("GET /subscription/", h.get)
	mux.HandleFunc("POST /subscription/webhook", h.webhookTest)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequestDto
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error(err.Error())
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		slog.Error(err.Error())
		http.Error(w, "bad request", http.StatusBadRequest) 
		return
	} 
	err := h.Service.Create(ctx, req)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "error to create subscription", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	var req GetSubscriptionRequestDto
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	sub, err := h.Service.Get(ctx, req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) webhookTest(w http.ResponseWriter, r *http.Request) {
	var j json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		slog.Error("err parse", err)
		return
	}
	slog.Info("req", j)
	json.NewEncoder(w).Encode("success")
}