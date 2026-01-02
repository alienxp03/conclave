package handlers

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alienxp03/conclave/web/app"
)

// serveSPA serves the React single-page application.
func (h *Handler) serveSPA(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Skip API and action routes
	if strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/debates") && r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	// Serve static files from dist
	distFS, err := fs.Sub(app.Dist, "dist")
	if err != nil {
		slog.Error("Failed to get dist filesystem", "error", err)
		http.NotFound(w, r)
		return
	}

	// Try to serve the file
	if path == "/" || path == "" {
		path = "/index.html"
	}

	// Remove leading slash for fs.Sub
	filePath := strings.TrimPrefix(path, "/")

	// Check if file exists
	if _, err := fs.Stat(distFS, filePath); err == nil {
		http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
		return
	}

	// If file not found, serve index.html (SPA routing)
	http.ServeFileFS(w, r, distFS, "index.html")
}

// RegisterSPARoutes registers routes for the React SPA.
// This should be called after API routes are registered.
func (h *Handler) RegisterSPARoutes(mux *http.ServeMux) {
	// Catch-all for SPA routing
	mux.HandleFunc("/", h.serveSPA)
}
