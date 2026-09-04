package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/opensource-easypanel/openpanel/frontend"
	"github.com/opensource-easypanel/openpanel/internal/adapter/db/sqlite"
	openpanelhttp "github.com/opensource-easypanel/openpanel/internal/adapter/http"
	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("OPENPANEL_DB")
	if dbPath == "" {
		dbPath = "data/openpanel.db"
	}

	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	ctx := context.Background()

	// Initialize Zero-CGO SQLite repository
	repo, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite repository: %v", err)
	}
	defer repo.Close()

	// Run versioned schema migrations
	if err := repo.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize Null Object / Docker adapter
	dockerAdapter := noop.NewNoOpDocker()

	// Embedded SPA Handler
	spaHandler := frontend.Handler()

	// Build HTTP Server with oRPC and SPA routes
	server := openpanelhttp.NewServer(openpanelhttp.ServerDependencies{
		DB:         repo,
		Docker:     dockerAdapter,
		SPAHandler: spaHandler,
	})

	addr := ":" + port
	fmt.Println("==================================================")
	fmt.Println("🚀 OpenSource Easypanel Control Plane")
	fmt.Printf("🌐 Serving Dashboard on: http://localhost:%s\n", port)
	fmt.Println("🔓 Pro Licensing: 100% Free & Unlocked")
	fmt.Println("🛡️ Telemetry: Zero Tracking & Phone-Home Free")
	fmt.Printf("📦 Embedded Assets: %t (Single-Binary Mode)\n", frontend.HasAssets())
	fmt.Println("==================================================")

	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
