package handlers

import (
	"net/http"

	"secure-document-transfer/internal/config"
	"secure-document-transfer/internal/database"
	"secure-document-transfer/internal/models"
)

// GetProfileHandler returns the authenticated user's profile
func GetProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user info from context (added by AuthMiddleware)
		userID := r.Context().Value("user_id")
		userEmail := r.Context().Value("user_email")

		if userID == nil || userEmail == nil {
			RespondWithError(w, http.StatusUnauthorized, "User not authenticated", "")
			return
		}

		// Get the full user data from the token
		authHeader := r.Header.Get("Authorization")
		tokenString := authHeader[7:] // Remove "Bearer " prefix

		user, err := config.SupabaseClient.Auth.WithToken(tokenString).GetUser()
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get user profile", err.Error())
			return
		}

		profile := models.User{
			ID:       user.User.ID.String(),
			Email:    user.User.Email,
			FullName: models.GetFullName(user.User.UserMetadata),
		}

		RespondWithJSON(w, http.StatusOK, profile)
	}
}

// SearchUsersHandler searches for users by email or name
func SearchUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get search query from URL parameters
		query := r.URL.Query().Get("q")
		if query == "" {
			RespondWithJSON(w, http.StatusOK, []models.User{})
			return
		}

		// Get current user ID from context to exclude them from results
		currentUserID := r.Context().Value("user_id")
		if currentUserID == nil {
			RespondWithError(w, http.StatusUnauthorized, "User not authenticated", "")
			return
		}

		// Search users from the database
		users, err := database.SearchUsers(query, currentUserID.(string))
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to search users", err.Error())
			return
		}

		RespondWithJSON(w, http.StatusOK, users)
	}
}

// GetUserPublicKeyHandler returns a specific user's public key
func GetUserPublicKeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from URL parameters
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			RespondWithError(w, http.StatusBadRequest, "user_id parameter is required", "")
			return
		}

		// Retrieve public key from database
		publicKey, err := database.GetUserPublicKey(userID)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, "User not found or has no public key", err.Error())
			return
		}

		RespondWithJSON(w, http.StatusOK, map[string]string{
			"user_id":    userID,
			"public_key": publicKey,
		})
	}
}

