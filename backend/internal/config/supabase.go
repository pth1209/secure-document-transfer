package config

import (
	"fmt"
	"os"

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

