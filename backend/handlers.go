package main

import (
	"encoding/json"
	"net/http"

	"github.com/supabase-community/gotrue-go/types"
)

// SignUpHandler handles user registration via Supabase Auth
func SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpRequest

		// Parse request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

	// Validate request
	if err := req.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	// Sign up with Supabase Auth
	signupRequest := types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
		Data: map[string]interface{}{
			"full_name": req.FullName,
		},
	}

	authResponse, err := SupabaseClient.Auth.Signup(signupRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create user", err.Error())
		return
	}

	// Respond with success
	response := SignUpResponse{
		Message:     "User registered successfully. Please check your email to verify your account.",
		AccessToken: authResponse.AccessToken,
		User: User{
			ID:       authResponse.User.ID.String(),
			Email:    authResponse.User.Email,
			FullName: getFullName(authResponse.User.UserMetadata),
		},
	}
		respondWithJSON(w, http.StatusCreated, response)
	}
}

// SignInHandler handles user login via Supabase Auth
func SignInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignInRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Sign in with Supabase Auth
	authResponse, err := SupabaseClient.Auth.SignInWithEmailPassword(req.Email, req.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", err.Error())
		return
	}

	// Respond with success
	response := SignInResponse{
		Message:      "Login successful",
		AccessToken:  authResponse.AccessToken,
		RefreshToken: authResponse.RefreshToken,
		User: User{
			ID:       authResponse.User.ID.String(),
			Email:    authResponse.User.Email,
			FullName: getFullName(authResponse.User.UserMetadata),
		},
	}
		respondWithJSON(w, http.StatusOK, response)
	}
}

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

// SignOutHandler handles user logout
func SignOutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Missing authorization header", "")
		return
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix

	err := SupabaseClient.Auth.WithToken(tokenString).Logout()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to sign out", err.Error())
		return
	}

		respondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Signed out successfully",
		})
	}
}

// Helper function to extract full_name from user metadata
func getFullName(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	if fullName, ok := metadata["full_name"].(string); ok {
		return fullName
	}
	return ""
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

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string, details string) {
	errorResponse := ErrorResponse{
		Error:   message,
		Details: details,
	}
	respondWithJSON(w, status, errorResponse)
}

