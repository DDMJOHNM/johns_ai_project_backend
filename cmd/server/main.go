package main

import (
	"context"
	"log"

	"github.com/jmason/john_ai_project/internal/router"
)

func main() {
	ctx := context.Background()

	// Create router
	r, err := router.NewRouter(ctx)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	// Start server with graceful shutdown
	if err := r.StartWithGracefulShutdown(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

