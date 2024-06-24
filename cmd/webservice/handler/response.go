package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type errorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

func WriteJson(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	anErr, ok := data.(error)
	if ok {
		data = errorResponse{ErrorMessage: anErr.Error()}
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		zerolog.Ctx(r.Context()).Err(err).Msg("Failed to write json data to response")
	}
}
