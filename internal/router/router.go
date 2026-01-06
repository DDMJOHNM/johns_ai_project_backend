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
	"github.com/jmason/john_ai_project/internal/logger"
	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

// Router handles HTTP routing
type Router struct {
	mux            *http.ServeMux
	server         *http.Server
	handler        *handler.ClientHandler
	cloudWatchLog  *logger.CloudWatchLogger
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

	// Create CloudWatch logger (optional - will work even if AWS credentials aren't available)
	var cwLogger *logger.CloudWatchLogger
	logGroupName := getEnv("CLOUDWATCH_LOG_GROUP", "/aws/ec2/john-ai-backend")
	logStreamName := getEnv("CLOUDWATCH_LOG_STREAM", "api-server")
	
	cwLogger, err = logger.NewCloudWatchLogger(ctx, logGroupName, logStreamName)
	if err != nil {
		log.Printf("Warning: CloudWatch logging not available: %v", err)
		// Continue anyway - local logging will still work
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

	// Specific routes first (must be registered before catch-all /api/clients/)
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

	// Catch-all route for /api/clients/{id}
	mux.HandleFunc("/api/clients/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Check if this is one of our specific routes (shouldn't happen, but safety check)
		if path == "/api/clients/active" || path == "/api/clients/inactive" || path == "/api/clients/add" {
			http.NotFound(w, r)
			return
		}

		// Only handle GET requests for client by ID
		if method == http.MethodGet {
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
			http.NotFound(w, r)
		}
	})

	// Middleware to log requests and strip stage prefix
	logAndStripHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		// Remove common stage prefixes if present
		if len(path) > 5 && path[:5] == "/prod" && path[5] == '/' {
			r.URL.Path = path[5:]
		} else if len(path) > 4 && path[:4] == "/dev" && path[4] == '/' {
			r.URL.Path = path[4:]
		} else if len(path) > 8 && path[:8] == "/staging" && path[8] == '/' {
			r.URL.Path = path[8:]
		}

		// Wrap response writer to capture status code
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the actual handler
		mux.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		if cwLogger != nil {
			go cwLogger.LogRequest(r.Context(), method, r.URL.Path, wrapped.statusCode, duration, r.RemoteAddr)
		}

		// Also log locally
		log.Printf("%s %s | Status: %d | Duration: %dms | Remote: %s", method, r.URL.Path, wrapped.statusCode, duration.Milliseconds(), r.RemoteAddr)
	})

	port := getEnv("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      logAndStripHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Router{
		mux:           mux,
		server:        server,
		handler:       clientHandler,
		cloudWatchLog: cwLogger,
	}, nil
}

// Start starts the HTTP server
func (r *Router) Start() error {
	log.Printf("Starting server on %s", r.server.Addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET /health - Health check")
	log.Printf("  GET /api/clients - Get all clients")
	log.Printf("  GET /api/clients/{id} - Get client by ID")
	log.Printf("  GET /api/clients/active - Get active clients")
	log.Printf("  GET /api/clients/inactive - Get inactive clients")
	log.Printf("  POST /api/clients/add - Create a new client")

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

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

