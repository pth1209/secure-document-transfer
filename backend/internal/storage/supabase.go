package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"secure-document-transfer/internal/config"
	storage_go "github.com/supabase-community/storage-go"
)

const (
	// BucketName is the name of the Supabase storage bucket for encrypted files
	BucketName = "encrypted-files"
)

// UploadEncryptedChunk uploads an encrypted chunk to Supabase Storage
// Returns the storage path of the uploaded chunk
// userToken is the JWT token of the authenticated user
func UploadEncryptedChunk(fileID string, chunkIndex int, data io.Reader, userToken string) (string, error) {
	// Generate storage path
	storagePath := fmt.Sprintf("%s/chunk_%d.enc", fileID, chunkIndex)

	// Read data into buffer
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, data); err != nil {
		return "", fmt.Errorf("failed to read chunk data: %w", err)
	}

	// Create a client with the user's token for authenticated storage access
	supabaseURL := os.Getenv("SUPABASE_URL")
	
	userClient := storage_go.NewClient(supabaseURL+"/storage/v1", userToken, nil)

	// Upload to Supabase Storage using the user's authenticated client
	_, err := userClient.UploadFile(BucketName, storagePath, buf)
	if err != nil {
		return "", fmt.Errorf("failed to upload chunk to storage: %w", err)
	}

	return storagePath, nil
}

// DownloadEncryptedChunk downloads an encrypted chunk from Supabase Storage
// userToken is the JWT token of the authenticated user
func DownloadEncryptedChunk(storagePath string, userToken string) ([]byte, error) {
	// Create a client with the user's token for authenticated storage access
	supabaseURL := os.Getenv("SUPABASE_URL")
	
	userClient := storage_go.NewClient(supabaseURL+"/storage/v1", userToken, nil)

	data, err := userClient.DownloadFile(BucketName, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download chunk from storage: %w", err)
	}

	return data, nil
}

// DeleteFile deletes all chunks of a file from Supabase Storage
func DeleteFile(fileID string) error {
	// List all files in the fileID folder
	files, err := config.SupabaseClient.Storage.ListFiles(BucketName, fileID, storage_go.FileSearchOptions{})
	if err != nil {
		return fmt.Errorf("failed to list file chunks: %w", err)
	}

	// Delete each file
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = fmt.Sprintf("%s/%s", fileID, file.Name)
	}

	if len(paths) > 0 {
		_, err = config.SupabaseClient.Storage.RemoveFile(BucketName, paths)
		if err != nil {
			return fmt.Errorf("failed to delete file chunks: %w", err)
		}
	}

	return nil
}

// InitializeBucket checks if the storage bucket exists
// This should be called once during application initialization
func InitializeBucket() error {
	// Check if bucket exists by attempting to list files in it
	_, err := config.SupabaseClient.Storage.ListFiles(BucketName, "", storage_go.FileSearchOptions{})
	if err != nil {
		// Bucket doesn't exist or we can't access it
		// Get service role key for bucket creation
		serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		if serviceRoleKey == "" {
			fmt.Printf("Warning: SUPABASE_SERVICE_ROLE_KEY is not set\n")
		}

		// Note: Bucket creation requires service role key and admin privileges
		// For now, we'll just log that the bucket needs to be created manually
		fmt.Printf("Warning: Bucket '%s' does not exist or is not accessible.\n", BucketName)
		fmt.Println("Please create it manually in Supabase Dashboard:")
		fmt.Println("  1. Go to Storage → New bucket")
		fmt.Println("  2. Name: encrypted-files")
		fmt.Println("  3. Public: false (private bucket)")
		fmt.Println("  4. File size limit: As needed")
		fmt.Println("  5. Allowed MIME types: all")
		fmt.Println()
	}

	return nil
}

