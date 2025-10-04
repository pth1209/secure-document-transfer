package main

import (
	"encoding/json"
	"net/http"
)

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
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

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string, details string) {
	errorResponse := ErrorResponse{
		Error:   message,
		Details: details,
	}
	respondWithJSON(w, status, errorResponse)
}

// getFullName extracts full_name from user metadata
func getFullName(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	if fullName, ok := metadata["full_name"].(string); ok {
		return fullName
	}
	return ""
}

