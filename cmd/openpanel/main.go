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
	"github.com/opensource-easypanel/openpanel/internal/adapter/docker"
	openpanelhttp "github.com/opensource-easypanel/openpanel/internal/adapter/http"
	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

func main() {
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
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

	// Initialize Docker adapter (with fallback to Null Object if Docker daemon unreachable)
	var dockerAdapter port.DockerPort
	liveDocker, err := docker.New()
	if err != nil {
		log.Printf("⚠️ Docker daemon unavailable (%v); falling back to NoOpDocker adapter", err)
		dockerAdapter = noop.NewNoOpDocker()
	} else {
		defer liveDocker.Close()
		dockerAdapter = liveDocker
		log.Println("🐳 Docker Engine & Swarm adapter initialized successfully")
	}

	// Embedded SPA Handler
	spaHandler := frontend.Handler()

	// Build HTTP Server with oRPC and SPA routes
	server := openpanelhttp.NewServer(openpanelhttp.ServerDependencies{
		DB:         repo,
		Docker:     dockerAdapter,
		SPAHandler: spaHandler,
	})

	addr := ":" + httpPort
	fmt.Println("==================================================")
	fmt.Println("🚀 OpenSource Easypanel Control Plane")
	fmt.Printf("🌐 Serving Dashboard on: http://localhost:%s\n", httpPort)
	fmt.Println("🔓 Pro Licensing: 100% Free & Unlocked")
	fmt.Println("🛡️ Telemetry: Zero Tracking & Phone-Home Free")
	fmt.Printf("📦 Embedded Assets: %t (Single-Binary Mode)\n", frontend.HasAssets())
	fmt.Println("==================================================")

	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
