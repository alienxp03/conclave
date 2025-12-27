package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alienxp03/dbate/internal/provider"
	"github.com/alienxp03/dbate/internal/storage"
	"github.com/alienxp03/dbate/web/handlers"
)

func main() {
	port := flag.Int("port", 8182, "Server port")
	dbPath := flag.String("db", "", "Database path (default: ~/.dbate/dbate.db)")
	flag.Parse()

	// Initialize storage
	path := *dbPath
	if path == "" {
		path = storage.DefaultDBPath()
	}

	store, err := storage.NewSQLiteStorage(path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize provider registry
	registry := provider.DefaultRegistry()

	// Create handler
	h := handlers.New(store, registry)

	// Setup routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Handle shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		server.Close()
	}()

	log.Printf("Starting dbate web server on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
