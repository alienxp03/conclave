package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alienxp03/conclave/internal/config"
	"github.com/alienxp03/conclave/internal/storage"
	"github.com/alienxp03/conclave/internal/workspace"
	"github.com/alienxp03/conclave/web/handlers"
)

func main() {
	port := flag.Int("port", 8182, "Server port")
	dbPath := flag.String("db", "", "Database path (default: ~/.conclave/conclave.db)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize slog
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if *debug {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	// Initialize storage
	path := *dbPath
	if path == "" {
		path = storage.DefaultDBPath()
	}

	slog.Info("Initializing storage", "path", path)
	store, err := storage.NewSQLiteStorage(path)
	if err != nil {
		slog.Error("Failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	if err := store.Initialize(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize provider registry
	cfg := config.Default()
	registry, err := cfg.CreateRegistry()
	if err != nil {
		slog.Error("Failed to initialize provider registry", "error", err)
		os.Exit(1)
	}

	// Initialize workspaces
	workspaces, err := workspace.NewManager()
	if err != nil {
		slog.Warn("Failed to initialize workspace manager", "error", err)
	}

	// Create handler
	h := handlers.New(store, registry, workspaces)

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
		slog.Info("Shutting down...")
		server.Close()
	}()

	slog.Info("Starting conclave web server", "url", fmt.Sprintf("http://localhost%s", addr))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}
