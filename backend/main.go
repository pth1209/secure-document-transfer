package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Supabase client
	if err := InitSupabaseClient(); err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Initialize database connection
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set global DB variable for handlers to use
	DB = db

	// Create router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Public routes (no authentication required)
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	api.HandleFunc("/signup", SignUpHandler()).Methods("POST")
	api.HandleFunc("/signin", SignInHandler()).Methods("POST")

	// Protected routes (authentication required)
	api.HandleFunc("/profile", AuthMiddleware(GetProfileHandler())).Methods("GET")
	api.HandleFunc("/signout", AuthMiddleware(SignOutHandler())).Methods("POST")
	api.HandleFunc("/users/search", AuthMiddleware(SearchUsersHandler())).Methods("GET")

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
