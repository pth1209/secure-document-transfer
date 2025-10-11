package handlers

import (
	"encoding/json"
	"net/http"

	"secure-document-transfer/internal/config"
	"secure-document-transfer/internal/crypto"
	"secure-document-transfer/internal/database"
	"secure-document-transfer/internal/models"

	"github.com/supabase-community/gotrue-go/types"
)

// SignUpHandler handles user registration via Supabase Auth
func SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SignUpRequest

		// Parse request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		// Validate request
		if err := req.Validate(); err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error(), "")
			return
		}

		// Generate encryption keys for the user
		// The private key is encrypted with a key derived from the user's password
		keys, err := crypto.GenerateUserKeys(req.Password)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to generate encryption keys", err.Error())
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

		authResponse, err := config.SupabaseClient.Auth.Signup(signupRequest)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Failed to create user", err.Error())
			return
		}

		// Create user record in database with encryption keys
		userID := authResponse.User.ID.String()
		err = database.CreateUser(userID, keys.PublicKeyPEM, keys.EncryptedPrivateKey, keys.Salt, keys.IV)
		if err != nil {
			// Note: User was created in Supabase Auth but failed to save keys to database
			RespondWithError(w, http.StatusInternalServerError, "Failed to save user encryption keys", err.Error())
			return
		}

		// Respond with success
		response := models.SignUpResponse{
			Message:     "User registered successfully. Please check your email to verify your account.",
			AccessToken: authResponse.AccessToken,
			User: models.User{
				ID:       authResponse.User.ID.String(),
				Email:    authResponse.User.Email,
				FullName: models.GetFullName(authResponse.User.UserMetadata),
			},
		}
		RespondWithJSON(w, http.StatusCreated, response)
	}
}

// SignInHandler handles user login via Supabase Auth
func SignInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SignInRequest

		// Parse request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		// Sign in with Supabase Auth
		authResponse, err := config.SupabaseClient.Auth.SignInWithEmailPassword(req.Email, req.Password)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Invalid credentials", err.Error())
			return
		}

		// Retrieve user encryption keys from database
		userID := authResponse.User.ID.String()
		keys, err := database.GetUserEncryptionKeys(userID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve encryption keys", err.Error())
			return
		}

		// Respond with success
		response := models.SignInResponse{
			Message:             "Login successful",
			AccessToken:         authResponse.AccessToken,
			RefreshToken:        authResponse.RefreshToken,
			User: models.User{
				ID:       userID,
				Email:    authResponse.User.Email,
				FullName: models.GetFullName(authResponse.User.UserMetadata),
			},
			EncryptedPrivateKey: keys.EncryptedPrivateKey,
			Salt:                keys.Salt,
			IV:                  keys.IV,
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

// SignOutHandler handles user logout
func SignOutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondWithError(w, http.StatusBadRequest, "Missing authorization header", "")
			return
		}

		tokenString := authHeader[7:] // Remove "Bearer " prefix

		err := config.SupabaseClient.Auth.WithToken(tokenString).Logout()
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to sign out", err.Error())
			return
		}

		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Signed out successfully",
		})
	}
}

