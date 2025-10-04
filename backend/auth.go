package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	supabase "github.com/supabase-community/supabase-go"
)

// SupabaseClient holds the Supabase client instance
var SupabaseClient *supabase.Client

// InitSupabaseClient initializes the Supabase client
func InitSupabaseClient() error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		return fmt.Errorf("failed to create Supabase client: %w", err)
	}

	SupabaseClient = client
	return nil
}

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
		user, err := SupabaseClient.Auth.WithToken(tokenString).GetUser()
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
				user, err := SupabaseClient.Auth.WithToken(tokenString).GetUser()
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

// VerifySupabaseJWT verifies a JWT token using Supabase's JWT secret
// This is an alternative method that validates the JWT locally without a network call
func VerifySupabaseJWT(tokenString string) (*jwt.Token, error) {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET is not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
