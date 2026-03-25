package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	mux           *http.ServeMux
	server        *http.Server
	handler       *handler.ClientHandler
	cloudWatchLog *logger.CloudWatchLogger
}

func NewRouter(ctx context.Context) (*Router, error) {
	dbClient, err := db.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB client: %w", err)
	}

	if err := dbClient.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DynamoDB: %w", err)
	}

	var cwLogger *logger.CloudWatchLogger
	logGroupName := getEnv("CLOUDWATCH_LOG_GROUP", "/aws/ec2/john-ai-backend")
	logStreamName := getEnv("CLOUDWATCH_LOG_STREAM", "api-server")

	cwLogger, err = logger.NewCloudWatchLogger(ctx, logGroupName, logStreamName)
	if err != nil {
		log.Printf("Warning: CloudWatch logging not available: %v", err)
	}

	// Setup repositories
	clientRepo := repository.NewClientRepository(dbClient.DynamoDB)
	userRepo := repository.NewUserRepository(dbClient.DynamoDB)

	// Setup services
	clientService := service.NewClientService(clientRepo)
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-CHANGE-IN-PRODUCTION-via-env-var")
	authService := service.NewAuthService(userRepo, jwtSecret)

	// Setup handlers
	clientHandler := handler.NewClientHandler(clientService)
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()

	// Public routes
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

	// Auth routes (public)
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Register(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Login(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/auth/me", authHandler.AuthMiddleware(authHandler.Me))

	// Protected API routes
	// IMPORTANT: More specific routes must be registered BEFORE less specific ones
	// because Go's ServeMux matches by longest prefix

	mux.HandleFunc("/api/clients/active", authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientHandler.GetActiveClients(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))

	mux.HandleFunc("/api/clients/inactive", authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientHandler.GetInactiveClients(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))

	mux.HandleFunc("/api/clients/add", authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTER] /api/clients/add handler - Method: %s, Path: %s", r.Method, r.URL.Path)
		fmt.Fprintf(os.Stderr, "[ROUTER] /api/clients/add handler - Method: %s, Path: %s\n", r.Method, r.URL.Path)
		if r.Method == http.MethodPost {
			log.Printf("[ROUTER] Calling CreateClient handler")
			fmt.Fprintf(os.Stderr, "[ROUTER] Calling CreateClient handler\n")
			clientHandler.CreateClient(w, r)
		} else {
			log.Printf("[ROUTER] Method not POST for /api/clients/add: %s", r.Method)
			fmt.Fprintf(os.Stderr, "[ROUTER] Method not POST for /api/clients/add: %s\n", r.Method)
			http.NotFound(w, r)
		}
	}))

	mux.HandleFunc("/api/clients/by-email", authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientHandler.GetClientByEmail(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))

	// PUT/PATCH /api/clients/update/{id} — alternate URL for client updates (must register before /api/clients/)
	const clientsUpdatePrefix = "/api/clients/update/"
	mux.HandleFunc(clientsUpdatePrefix, authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, clientsUpdatePrefix) {
			http.NotFound(w, r)
			return
		}
		id := strings.TrimSuffix(path[len(clientsUpdatePrefix):], "/")
		if id == "" || handler.ReservedClientPathID(id) {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPut && r.Method != http.MethodPatch {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), handler.ClientIDKey, id))
		clientHandler.UpdateClient(w, r)
	}))

	// Base route for GET /api/clients (protected)
	mux.HandleFunc("/api/clients", authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/clients" {
			clientHandler.GetClientList(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))

	// Catch-all route for /api/clients/{id} (protected)
	mux.HandleFunc("/api/clients/", authHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		log.Printf("[ROUTER] Catch-all /api/clients/ handler - Method: %s, Path: %s", method, path)
		fmt.Fprintf(os.Stderr, "[ROUTER] Catch-all /api/clients/ handler - Method: %s, Path: %s\n", method, path)

		if path == "/api/clients" || path == "/api/clients/" {
			http.NotFound(w, r)
			return
		}

		// Remove /api/clients/ prefix to get the segment (may be a reserved route or a client id).
		id := strings.TrimSuffix(path[len("/api/clients/"):], "/")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		// Depending on Go version / ServeMux, the subtree pattern /api/clients/ can also match
		// paths that have dedicated handlers. Delegate instead of 404 or treating as a client id.
		switch id {
		case "active":
			if method == http.MethodGet {
				clientHandler.GetActiveClients(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		case "inactive":
			if method == http.MethodGet {
				clientHandler.GetInactiveClients(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		case "add":
			if method == http.MethodPost {
				clientHandler.CreateClient(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		case "by-email":
			if method == http.MethodGet {
				clientHandler.GetClientByEmail(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		switch method {
		case http.MethodGet:
			r = r.WithContext(context.WithValue(r.Context(), handler.ClientIDKey, id))
			clientHandler.GetClientByID(w, r)
		case http.MethodPut, http.MethodPatch:
			r = r.WithContext(context.WithValue(r.Context(), handler.ClientIDKey, id))
			clientHandler.UpdateClient(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	// Middleware to log requests, strip stage prefix, and recover from panics
	logAndStripHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[MIDDLEWARE] PANIC recovered: %v", err)
				fmt.Fprintf(os.Stderr, "[MIDDLEWARE] PANIC recovered: %v\n", err)
				handler.RespondJSON(w, http.StatusInternalServerError, handler.ErrorResponse{
					Error:   "Internal server error",
					Message: "Request processing failed",
				})
			}
		}()

		start := time.Now()
		originalPath := r.URL.Path
		path := r.URL.Path
		method := r.Method

		// Log immediately to ensure we see all requests
		log.Printf("[MIDDLEWARE] Incoming request: %s %s (original: %s)", method, path, originalPath)
		fmt.Fprintf(os.Stderr, "[MIDDLEWARE] Incoming request: %s %s (original: %s)\n", method, path, originalPath)

		// Remove common stage prefixes if present
		if len(path) > 5 && path[:5] == "/prod" && path[5] == '/' {
			r.URL.Path = path[5:]
			log.Printf("[MIDDLEWARE] Stripped /prod prefix, new path: %s", r.URL.Path)
		} else if len(path) > 4 && path[:4] == "/dev" && path[4] == '/' {
			r.URL.Path = path[4:]
			log.Printf("[MIDDLEWARE] Stripped /dev prefix, new path: %s", r.URL.Path)
		} else if len(path) > 8 && path[:8] == "/staging" && path[8] == '/' {
			r.URL.Path = path[8:]
			log.Printf("[MIDDLEWARE] Stripped /staging prefix, new path: %s", r.URL.Path)
		}

		// Wrap response writer to capture status code
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		mux.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		log.Printf("[MIDDLEWARE] %s %s | Status: %d | Duration: %dms | Remote: %s", method, r.URL.Path, wrapped.statusCode, duration.Milliseconds(), r.RemoteAddr)
		fmt.Fprintf(os.Stderr, "[MIDDLEWARE] %s %s | Status: %d | Duration: %dms | Remote: %s\n", method, r.URL.Path, wrapped.statusCode, duration.Milliseconds(), r.RemoteAddr)

		if cwLogger != nil {
			go cwLogger.LogRequest(r.Context(), method, r.URL.Path, wrapped.statusCode, duration, r.RemoteAddr)
		}
	})

	port := getEnv("HTTP_PORT", "8081")
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

func (r *Router) Start() error {
	log.Printf("Starting server on %s", r.server.Addr)
	log.Printf("Available endpoints:")
	log.Printf("  Public:")
	log.Printf("    GET  /health - Health check")
	log.Printf("    POST /api/auth/register - Register new user")
	log.Printf("    POST /api/auth/login - Login (returns JWT token)")
	log.Printf("  Protected (requires Authorization: Bearer <token>):")
	log.Printf("    GET  /api/auth/me - Get current user info")
	log.Printf("    GET  /api/clients - Get all clients")
	log.Printf("    GET  /api/clients/{id} - Get client by ID")
	log.Printf("    GET  /api/clients/by-email?email=... - Get client by email")
	log.Printf("    PUT/PATCH /api/clients/{id} - Update a client")
	log.Printf("    PUT/PATCH /api/clients/update/{id} - Update a client (alternate path)")
	log.Printf("    GET  /api/clients/active - Get active clients")
	log.Printf("    GET  /api/clients/inactive - Get inactive clients")
	log.Printf("    POST /api/clients/add - Create a new client")

	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}

func (r *Router) StartWithGracefulShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

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
