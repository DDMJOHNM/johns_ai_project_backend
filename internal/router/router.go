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
	// IMPORTANT: More specific routes must be registered BEFORE less specific ones
	// because Go's ServeMux matches by longest prefix

	// Specific routes first
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
	mux.HandleFunc("/api/clients/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			clientHandler.CreateClient(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Base route for GET /api/clients
	mux.HandleFunc("/api/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/clients" {
			clientHandler.GetClientList(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Handle /api/clients/{id} pattern (must be last)
	// This catch-all route will match any /api/clients/* that isn't handled above
	mux.HandleFunc("/api/clients/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if this is one of our specific routes (shouldn't happen, but safety check)
		if path == "/api/clients/active" || path == "/api/clients/inactive" || path == "/api/clients/add" {
			// These should be handled by specific routes above, but if we get here, return 404
			http.NotFound(w, r)
			return
		}

		// Only handle GET requests for client by ID
		if r.Method == http.MethodGet {
			if path == "/api/clients" || path == "/api/clients/" {
				http.NotFound(w, r)
				return
			}
			// Remove /api/clients/ prefix to get the ID
			id := path[len("/api/clients/"):]
			if id == "" {
				http.NotFound(w, r)
				return
			}
			// Create a request with PathValue for the handler
			r = r.WithContext(context.WithValue(r.Context(), handler.ClientIDKey, id))
			clientHandler.GetClientByID(w, r)
		} else {
			// For non-GET requests, return 404 (POST /api/clients/add should be caught above)
			http.NotFound(w, r)
		}
	})

	// Middleware to strip stage prefix (e.g., /prod) from API Gateway requests
	// This handles the case where API Gateway forwards /prod/health instead of /health
	stripStagePrefixHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalPath := r.URL.Path
		// Remove common stage prefixes if present
		path := r.URL.Path
		if len(path) > 5 && path[:5] == "/prod" && len(path) > 5 && path[5] == '/' {
			r.URL.Path = path[5:] // Remove /prod
		} else if len(path) > 4 && path[:4] == "/dev" && len(path) > 4 && path[4] == '/' {
			r.URL.Path = path[4:] // Remove /dev
		} else if len(path) > 8 && path[:8] == "/staging" && len(path) > 8 && path[8] == '/' {
			r.URL.Path = path[8:] // Remove /staging
		}

		// Log all requests for debugging
		log.Printf("[%s] %s %s -> %s", r.Method, originalPath, r.RemoteAddr, r.URL.Path)

		mux.ServeHTTP(w, r)
	})

	port := getEnv("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      stripStagePrefixHandler,
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
	log.Printf("  POST /api/clients/add - Create a new client")
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
