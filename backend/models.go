package main

import (
	"regexp"
	"strings"
)

// User represents a user in the system (from Supabase Auth)
type User struct {
	ID       string `json:"id"`        // UUID from Supabase Auth
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// SignUpRequest represents the request body for user signup
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// SignUpResponse represents the response after successful signup
type SignUpResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token,omitempty"`
	User        User   `json:"user"`
}

// SignInRequest represents the request body for user signin
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignInResponse represents the response after successful signin
type SignInResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// Validate validates the signup request
func (req *SignUpRequest) Validate() error {
	// Trim whitespace
	req.Email = strings.TrimSpace(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)

	// Validate email
	if req.Email == "" {
		return &ValidationError{Field: "email", Message: "Email is required"}
	}
	if !isValidEmail(req.Email) {
		return &ValidationError{Field: "email", Message: "Invalid email format"}
	}

	// Validate password
	if req.Password == "" {
		return &ValidationError{Field: "password", Message: "Password is required"}
	}
	if len(req.Password) < 8 {
		return &ValidationError{Field: "password", Message: "Password must be at least 8 characters long"}
	}

	// Validate full name (optional but should not be too long)
	if len(req.FullName) > 255 {
		return &ValidationError{Field: "full_name", Message: "Full name is too long"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	// Basic email regex pattern
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

