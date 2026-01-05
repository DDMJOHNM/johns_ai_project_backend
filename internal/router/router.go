package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmason/john_ai_project/internal/db"
	"github.com/jmason/john_ai_project/internal/handler"
	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

// Router handles HTTP routing
type Router struct {
	mux     *http.ServeMux
	server  *http.Server
	handler *handler.ClientHandler
}

// NewRouter creates a new router with all routes configured
func NewRouter(ctx context.Context) (*Router, error) {
	// Create DynamoDB connection
	dbClient, err := db.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB client: %w", err)
	}

	// Test connection
	if err := dbClient.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DynamoDB: %w", err)
	}

	// Create repository
	clientRepo := repository.NewClientRepository(dbClient.DynamoDB)

	// Create service
	clientService := service.NewClientService(clientRepo)

	// Create handler
	clientHandler := handler.NewClientHandler(clientService)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.RespondJSON(w, http.StatusOK, handler.HealthResponse{
				Status:  "ok",
				Message: "Server is running",
			})
		} else {
			http.NotFound(w, r)
		}
	})

	// API routes
	mux.HandleFunc("/api/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/clients" {
			clientHandler.GetClientList(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/api/clients/active", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientHandler.GetActiveClients(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/api/clients/inactive", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientHandler.GetInactiveClients(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	// Handle /api/clients/{id} pattern
	mux.HandleFunc("/api/clients/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Extract ID from path: /api/clients/{id}
			path := r.URL.Path
			if path == "/api/clients" || path == "/api/clients/" {
				http.NotFound(w, r)
				return
			}
			// Remove /api/clients/ prefix to get the ID
			id := path[len("/api/clients/"):]
			if id == "active" || id == "inactive" {
				// These are handled by specific routes above
				http.NotFound(w, r)
				return
			}
			// Create a request with PathValue for the handler
			r = r.WithContext(context.WithValue(r.Context(), "client_id", id))
			clientHandler.GetClientByID(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/clients/add", func(w http.ResponseWriter, r *http.Request) {
		clientHandler.AddClient(w, r)
	})

	port := getEnv("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Router{
		mux:     mux,
		server:  server,
		handler: clientHandler,
	}, nil
}

// Start starts the HTTP server
func (r *Router) Start() error {
	log.Printf("Starting server on %s", r.server.Addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET /health - Health check")
	log.Printf("  GET /api/clients - Get all clients")
	log.Printf("  GET /api/clients/active - Get active clients")
	log.Printf("  GET /api/clients/inactive - Get inactive clients")
	log.Printf("  GET /api/clients/{id} - Get client by ID")
	log.Printf("  POST /api/clients/add - Add a new client")

	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}

// StartWithGracefulShutdown starts the server with graceful shutdown
func (r *Router) StartWithGracefulShutdown() error {
	// Channel to listen for interrupt or terminate signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := r.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := r.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
