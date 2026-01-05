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

	// API routes - handle /api/clients/* routes manually to avoid ServeMux prefix matching issues
	// Register both /api/clients (exact) and /api/clients/ (prefix) to catch all variations
	mux.HandleFunc("/api/clients", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Exact match: GET /api/clients (no trailing slash)
		if path == "/api/clients" && method == http.MethodGet {
			clientHandler.GetClientList(w, r)
			return
		}

		// If not exact match, let the /api/clients/ handler deal with it
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/clients/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		logMsg := fmt.Sprintf("DEBUG: /api/clients handler - Method: %s, Path: '%s', PathLen: %d", method, path, len(path))
		log.Printf(logMsg)
		fmt.Fprintf(os.Stderr, "%s\n", logMsg)

		// Exact match: GET /api/clients
		if path == "/api/clients" {
			if method == http.MethodGet {
				clientHandler.GetClientList(w, r)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		// Handle paths with /api/clients/ prefix
		prefixLen := len("/api/clients/")
		if len(path) > prefixLen {
			suffix := path[prefixLen:]
			logMsg = fmt.Sprintf("DEBUG: Extracted suffix: '%s' (path: '%s')", suffix, path)
			log.Printf(logMsg)
			fmt.Fprintf(os.Stderr, "%s\n", logMsg)

			switch suffix {
			case "active":
				if method == http.MethodGet {
					clientHandler.GetActiveClients(w, r)
				} else {
					http.NotFound(w, r)
				}
				return
			case "inactive":
				if method == http.MethodGet {
					clientHandler.GetInactiveClients(w, r)
				} else {
					http.NotFound(w, r)
				}
				return
			case "add":
				logMsg = fmt.Sprintf("DEBUG: /api/clients/add matched - Method: %s", method)
				log.Printf(logMsg)
				fmt.Fprintf(os.Stderr, "%s\n", logMsg)
				if method == http.MethodPost {
					log.Printf("DEBUG: Calling CreateClient handler")
					fmt.Fprintf(os.Stderr, "DEBUG: Calling CreateClient handler\n")
					clientHandler.CreateClient(w, r)
				} else {
					logMsg = fmt.Sprintf("DEBUG: Method not POST, returning 404. Method was: %s", method)
					log.Printf(logMsg)
					fmt.Fprintf(os.Stderr, "%s\n", logMsg)
					http.NotFound(w, r)
				}
				return
			default:
				// Handle /api/clients/{id} - only GET allowed
				if method == http.MethodGet {
					logMsg = fmt.Sprintf("DEBUG: Handling client by ID: %s", suffix)
					log.Printf(logMsg)
					fmt.Fprintf(os.Stderr, "%s\n", logMsg)
					r = r.WithContext(context.WithValue(r.Context(), handler.ClientIDKey, suffix))
					clientHandler.GetClientByID(w, r)
				} else {
					http.NotFound(w, r)
				}
				return
			}
		}

		// Fallback
		logMsg = "DEBUG: No match found, returning 404"
		log.Printf(logMsg)
		fmt.Fprintf(os.Stderr, "%s\n", logMsg)
		http.NotFound(w, r)
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

		// Log all requests for debugging (using both log and fmt for visibility)
		logMsg := fmt.Sprintf("[%s] %s %s -> %s", r.Method, originalPath, r.RemoteAddr, r.URL.Path)
		log.Printf(logMsg)
		fmt.Fprintf(os.Stderr, "%s\n", logMsg) // Also write to stderr for immediate visibility

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
