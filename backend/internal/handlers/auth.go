package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

// generateSecurePassword generates a random secure password
func generateSecurePassword() (string, error) {
	// Generate 16 random bytes (will be 24 characters in base64)
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Use base64 URL encoding (no special chars that might cause issues)
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateUserAndSendResetEmail creates a new user with a random password and sends a password reset email
func CreateUserAndSendResetEmail(email string) (*models.User, error) {
	// Generate a random secure password
	randomPassword, err := generateSecurePassword()
	if err != nil {
		return nil, err
	}

	// Generate encryption keys for the user
	keys, err := crypto.GenerateUserKeys(randomPassword)
	if err != nil {
		return nil, err
	}

	// Create user in Supabase Auth with auto-confirm disabled
	// Use the email as full_name by default
	signupRequest := types.SignupRequest{
		Email:    email,
		Password: randomPassword,
		Data: map[string]interface{}{
			"full_name": email, // Default to email as name
		},
	}

	authResponse, err := config.SupabaseClient.Auth.Signup(signupRequest)
	if err != nil {
		return nil, err
	}

	// Create user record in database with encryption keys
	userID := authResponse.User.ID.String()
	err = database.CreateUser(userID, keys.PublicKeyPEM, keys.EncryptedPrivateKey, keys.Salt, keys.IV)
	if err != nil {
		log.Printf("Warning: User created in auth but failed to save keys: %v", err)
		// Don't return error here, proceed to send reset email
	}

	// Send password reset email using Supabase REST API
	err = sendPasswordResetEmail(email)
	if err != nil {
		log.Printf("Warning: Failed to send password reset email to %s: %v", email, err)
		// Don't return error, user is created
	}

	return &models.User{
		ID:       userID,
		Email:    email,
		FullName: email,
	}, nil
}

// sendPasswordResetEmail sends a password reset email using Supabase's REST API
func sendPasswordResetEmail(email string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	frontendURL := os.Getenv("FRONTEND_URL")

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	// Default frontend URL if not set
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Prepare the request body with redirectTo parameter
	requestBody := map[string]string{
		"email":      email,
		"redirectTo": fmt.Sprintf("%s/reset-password", frontendURL),
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make the HTTP request to Supabase Auth API
	url := fmt.Sprintf("%s/auth/v1/recover", supabaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send password reset email, status: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("Password reset email sent successfully to %s with redirect to %s/reset-password", email, frontendURL)
	return nil
}

// RequestPasswordResetHandler handles password reset requests
func RequestPasswordResetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PasswordResetRequest

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

		// Send password reset email via Supabase
		err := sendPasswordResetEmail(req.Email)
		if err != nil {
			log.Printf("Failed to send password reset email: %v", err)
			// Don't reveal whether the email exists or not for security reasons
		}

		// Always return success to prevent email enumeration
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If an account with that email exists, a password reset link has been sent.",
		})
	}
}

// ResetPasswordHandler handles password reset with token
func ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PasswordResetConfirm

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

		// Update password using Supabase Auth API
		supabaseURL := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

		if supabaseURL == "" || supabaseKey == "" {
			RespondWithError(w, http.StatusInternalServerError, "Server configuration error", "")
			return
		}

		// Prepare the request body for Supabase
		requestBody := map[string]string{
			"password": req.NewPassword,
		}
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to process request", err.Error())
			return
		}

		// Make the HTTP request to Supabase Auth API to update password
		url := fmt.Sprintf("%s/auth/v1/user", supabaseURL)
		httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to create request", err.Error())
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("apikey", supabaseKey)
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", req.Token))

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to reset password", err.Error())
			return
		}
		defer resp.Body.Close()

		// Read the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to read response", err.Error())
			return
		}

		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to reset password, status: %d, body: %s", resp.StatusCode, string(body))
			RespondWithError(w, http.StatusBadRequest, "Invalid or expired reset token", string(body))
			return
		}

		log.Printf("Password reset successfully")
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Password reset successfully. You can now sign in with your new password.",
		})
	}
}

