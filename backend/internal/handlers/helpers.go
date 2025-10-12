package handlers

import (
	"encoding/json"
	"net/http"

	"secure-document-transfer/internal/models"
)

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to marshal response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, status int, message string, details string) {
	errorResponse := models.ErrorResponse{
		Error:   message,
		Details: details,
	}
	RespondWithJSON(w, status, errorResponse)
}

// parseJSON parses the request body into the given interface
func parseJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

