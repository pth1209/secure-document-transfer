package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"secure-document-transfer/internal/database"
	"secure-document-transfer/internal/storage"
)

// Mutex for synchronizing file metadata and recipient creation
// This prevents race conditions when chunks are uploaded concurrently
var (
	fileMetadataLock sync.Mutex
	fileMetadataMap  = make(map[string]bool) // tracks which files have had metadata created
)

	// SendFileChunkHandler handles encrypted file chunk uploads
func SendFileChunkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID := r.Context().Value("user_id")
		if userID == nil {
			RespondWithError(w, http.StatusUnauthorized, "User not authenticated", "")
			return
		}
		senderID := userID.(string)
		
		// Get user token for authenticated storage operations
		userToken := r.Context().Value("user_token")
		if userToken == nil {
			RespondWithError(w, http.StatusUnauthorized, "User token not found", "")
			return
		}
		token := userToken.(string)

		// Parse multipart form (10MB max memory)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Failed to parse form data", err.Error())
			return
		}

		// Extract form fields
		fileID := r.FormValue("file_id")
		chunkIndexStr := r.FormValue("chunk_index")
		totalChunksStr := r.FormValue("total_chunks")
		originalFilename := r.FormValue("original_filename")
		fileSizeStr := r.FormValue("file_size")
		chunkSizeStr := r.FormValue("chunk_size")
		iv := r.FormValue("iv")
		encryptedKeysJSON := r.FormValue("encrypted_keys")
		mimeType := r.FormValue("mime_type")

		// Validate required fields
		if fileID == "" || chunkIndexStr == "" || totalChunksStr == "" || originalFilename == "" || iv == "" || encryptedKeysJSON == "" {
			RespondWithError(w, http.StatusBadRequest, "Missing required fields", "")
			return
		}

		// Parse numeric values
		chunkIndex, err := strconv.Atoi(chunkIndexStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid chunk_index", err.Error())
			return
		}

		totalChunks, err := strconv.Atoi(totalChunksStr)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid total_chunks", err.Error())
			return
		}

		fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid file_size", err.Error())
			return
		}

		chunkSize, err := strconv.ParseInt(chunkSizeStr, 10, 64)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid chunk_size", err.Error())
			return
		}

		// Parse encrypted keys
		var encryptedKeys map[string]string
		if err := json.Unmarshal([]byte(encryptedKeysJSON), &encryptedKeys); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid encrypted_keys format", err.Error())
			return
		}

		// Get recipient emails
		recipientEmails := r.Form["recipient_emails[]"]
		if len(recipientEmails) == 0 {
			RespondWithError(w, http.StatusBadRequest, "At least one recipient email is required", "")
			return
		}

		// Get the encrypted file chunk
		file, _, err := r.FormFile("encrypted_chunk")
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Failed to get encrypted chunk", err.Error())
			return
		}
		defer file.Close()

		// Safely create file metadata and recipients (only once, even with concurrent requests)
		// This uses a mutex to ensure only one goroutine processes this for each file
		fileMetadataLock.Lock()
		if !fileMetadataMap[fileID] {
			// Create file metadata using safe upsert
			err = database.CreateFileMetadataIfNotExists(fileID, senderID, originalFilename, fileSize, totalChunks, mimeType)
			if err != nil {
				fileMetadataLock.Unlock()
				log.Printf("Error creating file metadata: %v", err)
				RespondWithError(w, http.StatusInternalServerError, "Failed to create file metadata", err.Error())
				return
			}

			// Process recipients
			var createdUsers []string
			var recipientRecords []struct {
				Email        string
				EncryptedKey string
				RecipientID  *string
			}

			for _, email := range recipientEmails {
				email = strings.TrimSpace(email)
				if email == "" {
					continue
				}

				// Check if user exists
				exists, err := database.UserExistsByEmail(email)
				if err != nil {
					fileMetadataLock.Unlock()
					log.Printf("Error checking if user exists for email %s: %v", email, err)
					RespondWithError(w, http.StatusInternalServerError, "Failed to check user existence", err.Error())
					return
				}

				// If user doesn't exist, create them and send password reset email
				var recipientID *string
				if !exists {
					log.Printf("Creating new user for email: %s", email)
					newUser, err := CreateUserAndSendResetEmail(email)
					if err != nil {
						fileMetadataLock.Unlock()
						log.Printf("Failed to create user for email %s: %v", email, err)
						RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err.Error())
						return
					}
					createdUsers = append(createdUsers, newUser.Email)
					log.Printf("Successfully created user %s and sent password reset email", newUser.Email)
					recipientID = &newUser.ID
				} else {
					// Get existing user ID
					user, err := database.GetUserByEmail(email)
					if err == nil {
						recipientID = &user.ID
					}
				}

				// Get encrypted key for this recipient
				encryptedKey, hasKey := encryptedKeys[email]
				if !hasKey {
					log.Printf("Warning: No encrypted key provided for recipient %s", email)
					encryptedKey = ""
				}

				recipientRecords = append(recipientRecords, struct {
					Email        string
					EncryptedKey string
					RecipientID  *string
				}{
					Email:        email,
					EncryptedKey: encryptedKey,
					RecipientID:  recipientID,
				})
			}

			// Create all recipient records in a transaction
			if len(recipientRecords) > 0 {
				err = database.CreateFileRecipientsIfNotExists(fileID, recipientRecords)
				if err != nil {
					fileMetadataLock.Unlock()
					log.Printf("Error creating file recipient records: %v", err)
					RespondWithError(w, http.StatusInternalServerError, "Failed to store recipient info", err.Error())
					return
				}
			}

			// Mark that we've processed metadata for this file
			fileMetadataMap[fileID] = true

			// Include created users in log
			if len(createdUsers) > 0 {
				log.Printf("Created new users: %v", createdUsers)
			}

			log.Printf("Created file metadata and recipients for file ID: %s", fileID)
		}
		fileMetadataLock.Unlock()

		// Upload encrypted chunk to Supabase Storage
		// This can happen concurrently for different chunks
		storagePath, err := storage.UploadEncryptedChunk(fileID, chunkIndex, file, token)
		if err != nil {
			log.Printf("Error uploading chunk to storage: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to upload chunk", err.Error())
			return
		}

		// Store chunk metadata in database
		// PostgreSQL handles concurrent inserts well
		err = database.CreateFileChunk(fileID, chunkIndex, chunkSize, storagePath, iv)
		if err != nil {
			log.Printf("Error creating chunk metadata: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to store chunk metadata", err.Error())
			return
		}

		// Check if this is the last chunk and mark file as complete
		// Multiple chunks might think they're last due to concurrency, but the update is idempotent
		if chunkIndex == totalChunks-1 {
			err = database.MarkFileComplete(fileID)
			if err != nil {
				log.Printf("Error marking file as complete: %v", err)
				// Don't fail the request, just log it
			}
			log.Printf("File %s marked as complete", fileID)
		}

		log.Printf("Stored encrypted chunk - File ID: %s, Chunk: %d/%d, Filename: %s, Storage: %s",
			fileID, chunkIndex+1, totalChunks, originalFilename, storagePath)

		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message":       "Encrypted chunk uploaded successfully",
			"file_id":       fileID,
			"chunk_index":   chunkIndex,
			"total_chunks":  totalChunks,
			"storage_path":  storagePath,
		})
	}
}

