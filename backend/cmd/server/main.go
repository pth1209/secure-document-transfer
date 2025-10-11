package main

import (
	"log"
	"net/http"
	"os"

	"secure-document-transfer/internal/config"
	"secure-document-transfer/internal/database"
	"secure-document-transfer/internal/handlers"
	"secure-document-transfer/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Supabase client
	if err := config.InitSupabaseClient(); err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set global DB variable for handlers to use
	database.DB = db

	// Create router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Public routes (no authentication required)
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	api.HandleFunc("/signup", handlers.SignUpHandler()).Methods("POST")
	api.HandleFunc("/signin", handlers.SignInHandler()).Methods("POST")
	api.HandleFunc("/password-reset/request", handlers.RequestPasswordResetHandler()).Methods("POST")
	api.HandleFunc("/password-reset/reset", handlers.ResetPasswordHandler()).Methods("POST")

	// Protected routes (authentication required)
	api.HandleFunc("/profile", middleware.AuthMiddleware(handlers.GetProfileHandler())).Methods("GET")
	api.HandleFunc("/signout", middleware.AuthMiddleware(handlers.SignOutHandler())).Methods("POST")
	api.HandleFunc("/users/search", middleware.AuthMiddleware(handlers.SearchUsersHandler())).Methods("GET")
	api.HandleFunc("/users/public-key", middleware.AuthMiddleware(handlers.GetUserPublicKeyHandler())).Methods("GET")
	api.HandleFunc("/files/send-chunk", middleware.AuthMiddleware(handlers.SendFileChunkHandler())).Methods("POST")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Printf("Supabase Auth integration enabled")
	if err := http.ListenAndServe(":"+port, enableCORS(router)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// enableCORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HealthCheckHandler returns the health status of the API
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

