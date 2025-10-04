package main

import (
	"net/http"
)

// GetProfileHandler returns the authenticated user's profile
func GetProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user info from context (added by AuthMiddleware)
		userID := r.Context().Value("user_id")
		userEmail := r.Context().Value("user_email")

		if userID == nil || userEmail == nil {
			respondWithError(w, http.StatusUnauthorized, "User not authenticated", "")
			return
		}

		// Get the full user data from the token
		authHeader := r.Header.Get("Authorization")
		tokenString := authHeader[7:] // Remove "Bearer " prefix

		user, err := SupabaseClient.Auth.WithToken(tokenString).GetUser()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get user profile", err.Error())
			return
		}

		profile := User{
			ID:       user.User.ID.String(),
			Email:    user.User.Email,
			FullName: getFullName(user.User.UserMetadata),
		}

		respondWithJSON(w, http.StatusOK, profile)
	}
}

// SearchUsersHandler searches for users by email or name
func SearchUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get search query from URL parameters
		query := r.URL.Query().Get("q")
		if query == "" {
			respondWithJSON(w, http.StatusOK, []User{})
			return
		}

		// Get current user ID from context to exclude them from results
		currentUserID := r.Context().Value("user_id")
		if currentUserID == nil {
			respondWithError(w, http.StatusUnauthorized, "User not authenticated", "")
			return
		}

		// Search users from the database
		// This queries the auth.users table to find users matching the search query
		users, err := searchUsersFromDatabase(query, currentUserID.(string))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to search users", err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, users)
	}
}

// searchUsersFromDatabase searches for users in the Supabase auth schema
func searchUsersFromDatabase(query string, excludeUserID string) ([]User, error) {
	// Build SQL query to search users by email or full_name
	// The auth.users table contains user metadata
	sqlQuery := `
		SELECT 
			id::text,
			email,
			COALESCE(raw_user_meta_data->>'full_name', '') as full_name
		FROM auth.users
		WHERE 
			id::text != $1
			AND (
				LOWER(email) LIKE LOWER($2)
				OR LOWER(COALESCE(raw_user_meta_data->>'full_name', '')) LIKE LOWER($2)
			)
		ORDER BY email
		LIMIT 20
	`

	searchPattern := "%" + query + "%"

	rows, err := DB.Query(sqlQuery, excludeUserID, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.FullName); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty array instead of nil if no users found
	if users == nil {
		users = []User{}
	}

	return users, nil
}

