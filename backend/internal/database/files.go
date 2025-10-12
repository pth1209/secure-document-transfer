package database

import (
	"database/sql"
	"fmt"
)

// FileMetadata represents file metadata in the database
type FileMetadata struct {
	ID               string
	FileID           string
	SenderID         string
	OriginalFilename string
	FileSize         int64
	TotalChunks      int
	MimeType         sql.NullString
}

// FileChunk represents a file chunk in the database
type FileChunk struct {
	ID           string
	FileID       string
	ChunkIndex   int
	ChunkSize    int64
	StoragePath  string
	EncryptionIV string
}

// FileRecipient represents a file recipient and their encrypted key
type FileRecipient struct {
	ID               string
	FileID           string
	RecipientID      sql.NullString
	RecipientEmail   string
	EncryptedFileKey string
}

// CreateFileMetadata creates a new file metadata record
func CreateFileMetadata(fileID, senderID, originalFilename string, fileSize int64, totalChunks int, mimeType string) error {
	query := `
		INSERT INTO public.file_metadata (file_id, sender_id, original_filename, file_size, total_chunks, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	var mimeTypeVal sql.NullString
	if mimeType != "" {
		mimeTypeVal = sql.NullString{String: mimeType, Valid: true}
	}

	_, err := DB.Exec(query, fileID, senderID, originalFilename, fileSize, totalChunks, mimeTypeVal)
	if err != nil {
		return fmt.Errorf("failed to create file metadata: %w", err)
	}

	return nil
}

// CreateFileChunk creates a new file chunk record
func CreateFileChunk(fileID string, chunkIndex int, chunkSize int64, storagePath, encryptionIV string) error {
	query := `
		INSERT INTO public.file_chunks (file_id, chunk_index, chunk_size, storage_path, encryption_iv)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := DB.Exec(query, fileID, chunkIndex, chunkSize, storagePath, encryptionIV)
	if err != nil {
		return fmt.Errorf("failed to create file chunk: %w", err)
	}

	return nil
}

// CreateFileRecipient creates a new file recipient record
func CreateFileRecipient(fileID, recipientEmail, encryptedFileKey string, recipientID *string) error {
	query := `
		INSERT INTO public.file_recipients (file_id, recipient_id, recipient_email, encrypted_file_key)
		VALUES ($1, $2, $3, $4)
	`

	var recipientIDVal sql.NullString
	if recipientID != nil && *recipientID != "" {
		recipientIDVal = sql.NullString{String: *recipientID, Valid: true}
	}

	_, err := DB.Exec(query, fileID, recipientIDVal, recipientEmail, encryptedFileKey)
	if err != nil {
		return fmt.Errorf("failed to create file recipient: %w", err)
	}

	return nil
}

// FileMetadataExists checks if file metadata already exists
func FileMetadataExists(fileID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM public.file_metadata WHERE file_id = $1)`

	var exists bool
	err := DB.QueryRow(query, fileID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check file metadata existence: %w", err)
	}

	return exists, nil
}

// MarkFileComplete marks a file as completely uploaded
func MarkFileComplete(fileID string) error {
	query := `UPDATE public.file_metadata SET completed_at = NOW() WHERE file_id = $1`

	_, err := DB.Exec(query, fileID)
	if err != nil {
		return fmt.Errorf("failed to mark file as complete: %w", err)
	}

	return nil
}

// CreateFileMetadataIfNotExists creates file metadata only if it doesn't exist yet
// This is safe for concurrent chunk uploads
func CreateFileMetadataIfNotExists(fileID, senderID, originalFilename string, fileSize int64, totalChunks int, mimeType string) error {
	// Use INSERT ... ON CONFLICT DO NOTHING for safe concurrent inserts
	query := `
		INSERT INTO public.file_metadata (file_id, sender_id, original_filename, file_size, total_chunks, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (file_id) DO NOTHING
	`

	var mimeTypeVal sql.NullString
	if mimeType != "" {
		mimeTypeVal = sql.NullString{String: mimeType, Valid: true}
	}

	_, err := DB.Exec(query, fileID, senderID, originalFilename, fileSize, totalChunks, mimeTypeVal)
	if err != nil {
		return fmt.Errorf("failed to create file metadata: %w", err)
	}

	return nil
}

// CreateFileRecipientsIfNotExists creates file recipient records if they don't exist yet
// This is safe for concurrent chunk uploads
func CreateFileRecipientsIfNotExists(fileID string, recipients []struct {
	Email        string
	EncryptedKey string
	RecipientID  *string
}) error {
	// Start a transaction for atomic inserts
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO public.file_recipients (file_id, recipient_id, recipient_email, encrypted_file_key)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (file_id, recipient_email) DO NOTHING
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, recipient := range recipients {
		var recipientIDVal sql.NullString
		if recipient.RecipientID != nil && *recipient.RecipientID != "" {
			recipientIDVal = sql.NullString{String: *recipient.RecipientID, Valid: true}
		}

		_, err = stmt.Exec(fileID, recipientIDVal, recipient.Email, recipient.EncryptedKey)
		if err != nil {
			return fmt.Errorf("failed to create file recipient: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

