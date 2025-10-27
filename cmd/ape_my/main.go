package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ticktockbent/ape_my/internal/cli"
	"github.com/ticktockbent/ape_my/internal/schema"
	"github.com/ticktockbent/ape_my/internal/server"
	"github.com/ticktockbent/ape_my/internal/storage"
)

func main() {
	// Parse command line arguments
	config, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		cli.PrintHelp()
		os.Exit(1)
	}

	// Handle help flag
	if config.ShowHelp {
		cli.PrintHelp()
		os.Exit(0)
	}

	// Handle version flag
	if config.ShowVersion {
		cli.PrintVersion()
		os.Exit(0)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print configuration
	fmt.Fprintf(os.Stderr, "ape_my v%s\n", cli.Version)
	fmt.Fprintf(os.Stderr, "Configuration: %s\n\n", config.String())

	// Phase 2: Load and parse schema
	log.Println("Loading schema...")
	loader := schema.NewLoader()
	if err := loader.LoadFromFile(config.SchemaFile); err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	entityNames := loader.GetEntityNames()
	log.Printf("Loaded %d entities: %v", len(entityNames), entityNames)

	// Build route map
	routeMap, err := loader.BuildRouteMap()
	if err != nil {
		log.Fatalf("Failed to build route map: %v", err)
	}

	// Phase 3: Initialize storage
	log.Println("Initializing storage...")
	store := storage.NewInMemoryStore()
	if err := store.Initialize(entityNames); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Load seed data if provided
	if config.SeedFile != "" {
		log.Printf("Loading seed data from %s...", config.SeedFile)
		seedData, err := schema.LoadSeedData(config.SeedFile)
		if err != nil {
			log.Fatalf("Failed to load seed data: %v", err)
		}

		// Validate seed data against schema
		if err := loader.ValidateSeedData(seedData); err != nil {
			log.Fatalf("Seed data validation failed: %v", err)
		}

		// Load seed data into storage
		for entityName, entities := range seedData {
			if err := store.Seed(entityName, entities); err != nil {
				log.Fatalf("Failed to seed %s: %v", entityName, err)
			}
			log.Printf("Seeded %d %s", len(entities), entityName)
		}
	}

	// Phase 4: Start HTTP server
	srv := server.New(config.Port, store, routeMap, loader)
	srv.RegisterRoutes()

	log.Printf("\n=== Ape_my is ready! ===")
	log.Printf("API endpoints available:")
	for _, route := range routeMap.GetRoutes() {
		log.Printf("  - %s (GET, POST)", route.CollectionPath)
		log.Printf("  - %s/<id> (GET, PUT, PATCH, DELETE)", route.CollectionPath)
	}
	log.Println()

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
