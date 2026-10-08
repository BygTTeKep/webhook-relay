package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Wrap(log *zap.Logger, h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			apiErr = &APIError{
				Status:  http.StatusInternalServerError,
				Message: "Internal server error",
				Err:     err,
			}
		}

		fields := []zap.Field{
			zap.String("Method", r.Method),
			zap.String("URL", r.URL.Path),
			zap.Int("Status", apiErr.Status),
			zap.Error(err),
		}
		if apiErr.Status >= 500 {
			log.Error("request failed", fields...)
		} else {
			log.Error("request rejected", fields...)
		}
		writeJSON(w, apiErr.Status, map[string]string{"error": apiErr.Message})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
