package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"secure-document-transfer/internal/database"
)

// SendFileChunkHandler handles file chunk uploads
func SendFileChunkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// Validate required fields
		if fileID == "" || chunkIndexStr == "" || totalChunksStr == "" || originalFilename == "" {
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

		// Get recipient emails
		recipientEmails := r.Form["recipient_emails[]"]
		if len(recipientEmails) == 0 {
			RespondWithError(w, http.StatusBadRequest, "At least one recipient email is required", "")
			return
		}

		// Get the file chunk
		file, _, err := r.FormFile("chunk")
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, "Failed to get file chunk", err.Error())
			return
		}
		defer file.Close()

		// Process each recipient email
		var createdUsers []string
		for _, email := range recipientEmails {
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}

			// Check if user exists
			exists, err := database.UserExistsByEmail(email)
			if err != nil {
				log.Printf("Error checking if user exists for email %s: %v", email, err)
				RespondWithError(w, http.StatusInternalServerError, "Failed to check user existence", err.Error())
				return
			}

			// If user doesn't exist, create them and send password reset email
			if !exists {
				log.Printf("Creating new user for email: %s", email)
				newUser, err := CreateUserAndSendResetEmail(email)
				if err != nil {
					log.Printf("Failed to create user for email %s: %v", email, err)
					RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err.Error())
					return
				}
				createdUsers = append(createdUsers, newUser.Email)
				log.Printf("Successfully created user %s and sent password reset email", newUser.Email)
			}
		}

		// TODO: Store the file chunk in your storage system (e.g., Supabase Storage)
		// For now, just log the chunk info
		log.Printf("Received file chunk - File ID: %s, Chunk: %d/%d, Filename: %s, Size: %d/%d",
			fileID, chunkIndex+1, totalChunks, originalFilename, chunkSize, fileSize)

		// Prepare response message
		message := "File chunk uploaded successfully"
		if len(createdUsers) > 0 {
			message += ". Created new users and sent password reset emails to: " + strings.Join(createdUsers, ", ")
		}

		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message":       message,
			"created_users": createdUsers,
			"file_id":       fileID,
			"chunk_index":   chunkIndex,
			"total_chunks":  totalChunks,
		})
	}
}

