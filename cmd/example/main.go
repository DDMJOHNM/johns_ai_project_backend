package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jmason/john_ai_project/internal/db"
	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

func main() {
	ctx := context.Background()

	// Create DynamoDB connection
	fmt.Println("Connecting to DynamoDB...")
	dbClient, err := db.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create DB client: %v", err)
	}

	// Test connection
	if err := dbClient.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping DynamoDB: %v", err)
	}
	fmt.Printf("✓ Connected to DynamoDB at %s (region: %s)\n", dbClient.Endpoint, dbClient.Region)

	// Create repository
	clientRepo := repository.NewClientRepository(dbClient.DynamoDB)

	// Create service
	clientService := service.NewClientService(clientRepo)

	// Get all clients
	fmt.Println("\nFetching all clients...")
	clients, err := clientService.GetClientList(ctx)
	if err != nil {
		log.Fatalf("Failed to get client list: %v", err)
	}

	fmt.Printf("✓ Found %d clients\n\n", len(clients))

	// Display clients
	if len(clients) > 0 {
		fmt.Println("Client List:")
		fmt.Println("=" + string(make([]byte, 80)))
		for i, client := range clients {
			fmt.Printf("\n%d. %s %s\n", i+1, client.FirstName, client.LastName)
			fmt.Printf("   ID: %s\n", client.ID)
			fmt.Printf("   Email: %s\n", client.Email)
			fmt.Printf("   Phone: %s\n", client.Phone)
			fmt.Printf("   Status: %s\n", client.Status)
		}
		fmt.Println("\n" + string(make([]byte, 80)))
	}

	// Get active clients
	fmt.Println("\nFetching active clients...")
	activeClients, err := clientService.GetActiveClients(ctx)
	if err != nil {
		log.Fatalf("Failed to get active clients: %v", err)
	}
	fmt.Printf("✓ Found %d active clients\n", len(activeClients))

	// Output as JSON if requested
	if len(os.Args) > 1 && os.Args[1] == "--json" {
		jsonOutput, err := json.MarshalIndent(clients, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		fmt.Println("\nJSON Output:")
		fmt.Println(string(jsonOutput))
	}
}

