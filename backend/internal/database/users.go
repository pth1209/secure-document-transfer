package database

import (
	"fmt"

	"github.com/lib/pq"
	"secure-document-transfer/internal/models"
)

// CreateUser creates a new user record in the public.users table
// This stores the user's encryption keys (public key and encrypted private key)
func CreateUser(userID, publicKey, encryptedPrivateKey, salt, iv string) error {
	query := `
		INSERT INTO public.users (id, public_key, encrypted_private_key, salt, iv, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	
	_, err := DB.Exec(query, userID, publicKey, encryptedPrivateKey, salt, iv)
	if err != nil {
		return fmt.Errorf("failed to create user record: %w", err)
	}
	
	return nil
}

// GetUserEncryptionKeys retrieves a user's encryption keys from the database
func GetUserEncryptionKeys(userID string) (*models.UserEncryptionKeys, error) {
	query := `
		SELECT public_key, encrypted_private_key, salt, iv
		FROM public.users
		WHERE id = $1
	`
	
	var keys models.UserEncryptionKeys
	err := DB.QueryRow(query, userID).Scan(
		&keys.PublicKey,
		&keys.EncryptedPrivateKey,
		&keys.Salt,
		&keys.IV,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user encryption keys: %w", err)
	}
	
	return &keys, nil
}

// GetUserPublicKey retrieves only the public key for a specific user
func GetUserPublicKey(userID string) (string, error) {
	query := `
		SELECT public_key
		FROM public.users
		WHERE id = $1
	`
	
	var publicKey string
	err := DB.QueryRow(query, userID).Scan(&publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve user public key: %w", err)
	}
	
	return publicKey, nil
}

// GetUserByEmail retrieves a user by their email address
func GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT 
			id::text,
			email,
			COALESCE(raw_user_meta_data->>'full_name', '') as full_name
		FROM auth.users
		WHERE LOWER(email) = LOWER($1)
	`
	
	var user models.User
	err := DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.FullName)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	return &user, nil
}

// UserExistsByEmail checks if a user exists by email
func UserExistsByEmail(email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM auth.users WHERE LOWER(email) = LOWER($1)
		)
	`
	
	var exists bool
	err := DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	return exists, nil
}

// SearchUsers searches for users in the Supabase auth schema
func SearchUsers(query string, excludeUserID string) ([]models.User, error) {
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

	var users []models.User
	for rows.Next() {
		var user models.User
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
		users = []models.User{}
	}

	return users, nil
}

// GetPublicKeysByEmails retrieves public keys for multiple email addresses
// Returns a map of email -> public_key for users that exist and have keys
func GetPublicKeysByEmails(emails []string) (map[string]string, error) {
	if len(emails) == 0 {
		return map[string]string{}, nil
	}

	// Convert emails to lowercase for case-insensitive matching
	lowercaseEmails := make([]string, len(emails))
	for i, email := range emails {
		lowercaseEmails[i] = email
	}

	query := `
		SELECT 
			au.email,
			pu.public_key
		FROM auth.users au
		INNER JOIN public.users pu ON au.id = pu.id
		WHERE LOWER(au.email) = ANY($1)
	`

	rows, err := DB.Query(query, pq.Array(lowercaseEmails))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve public keys: %w", err)
	}
	defer rows.Close()

	publicKeys := make(map[string]string)
	for rows.Next() {
		var email, publicKey string
		if err := rows.Scan(&email, &publicKey); err != nil {
			return nil, fmt.Errorf("failed to scan public key: %w", err)
		}
		publicKeys[email] = publicKey
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating public keys: %w", err)
	}

	return publicKeys, nil
}

