package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"secure-document-transfer/internal/config"
	"secure-document-transfer/internal/models"
)

// AuthMiddleware verifies the JWT token from Supabase Auth
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization header", "")
			return
		}

		// Extract the token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format", "")
			return
		}

		tokenString := parts[1]

		// Verify the token with Supabase
		user, err := config.SupabaseClient.Auth.WithToken(tokenString).GetUser()
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", user.User.ID.String())
		ctx = context.WithValue(ctx, "user_email", user.User.Email)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't require authentication
// It adds user info to context if token is present and valid
func OptionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				user, err := config.SupabaseClient.Auth.WithToken(tokenString).GetUser()
				if err == nil {
					ctx := context.WithValue(r.Context(), "user_id", user.User.ID.String())
					ctx = context.WithValue(ctx, "user_email", user.User.Email)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string, details string) {
	errorResponse := models.ErrorResponse{
		Error:   message,
		Details: details,
	}
	respondWithJSON(w, status, errorResponse)
}

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

