package api

import (
	"encoding/json"
	"net/http"

	"kakeibo/internal/service/api/dto"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, dto.ErrorResponse{
		Error: msg,
	})
}

func writeSuccess(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, dto.SuccessResponse{
		Status: msg,
	})
}
