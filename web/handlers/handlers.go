// Package handlers provides HTTP handlers for the web interface.
package handlers

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alienxp03/dbate/internal/core"
	"github.com/alienxp03/dbate/internal/engine"
	"github.com/alienxp03/dbate/internal/export"
	"github.com/alienxp03/dbate/internal/persona"
	"github.com/alienxp03/dbate/internal/provider"
	"github.com/alienxp03/dbate/internal/storage"
	"github.com/alienxp03/dbate/internal/style"
)

//go:embed templates/*.html
var templateFS embed.FS

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	engine    *engine.Engine
	registry  *provider.Registry
	storage   storage.Storage
	templates *template.Template
}

// New creates a new Handler.
func New(store storage.Storage, registry *provider.Registry) *Handler {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 3:04 PM")
		},
		"shortID": func(id string) string {
			if len(id) > 8 {
				return id[:8]
			}
			return id
		},
		"truncate": func(s string, n int) string {
			if len(s) > n {
				return s[:n-3] + "..."
			}
			return s
		},
		"statusColor": func(status core.DebateStatus) string {
			switch status {
			case core.StatusCompleted:
				return "bg-green-100 text-green-800"
			case core.StatusInProgress:
				return "bg-blue-100 text-blue-800"
			case core.StatusFailed:
				return "bg-red-100 text-red-800"
			default:
				return "bg-gray-100 text-gray-800"
			}
		},
		"nl2br": func(s string) template.HTML {
			escaped := template.HTMLEscapeString(s)
			return template.HTML(strings.ReplaceAll(escaped, "\n", "<br>"))
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	return &Handler{
		engine:    engine.New(store, registry),
		registry:  registry,
		storage:   store,
		templates: tmpl,
	}
}

// RegisterRoutes registers all HTTP routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Pages
	mux.HandleFunc("GET /", h.handleIndex)
	mux.HandleFunc("GET /debates", h.handleDebatesList)
	mux.HandleFunc("GET /debates/{id}", h.handleDebateView)
	mux.HandleFunc("GET /new", h.handleNewDebateForm)

	// HTMX partials
	mux.HandleFunc("GET /partials/debates", h.handleDebatesPartial)
	mux.HandleFunc("GET /partials/debate/{id}", h.handleDebatePartial)
	mux.HandleFunc("GET /partials/debate/{id}/turns", h.handleDebateTurnsPartial)

	// Actions
	mux.HandleFunc("POST /debates", h.handleCreateDebate)
	mux.HandleFunc("POST /debates/{id}/run", h.handleRunDebate)
	mux.HandleFunc("POST /debates/{id}/next", h.handleNextTurn)
	mux.HandleFunc("POST /debates/{id}/lock", h.handleLockDebate)
	mux.HandleFunc("POST /debates/{id}/unlock", h.handleUnlockDebate)
	mux.HandleFunc("DELETE /debates/{id}", h.handleDeleteDebate)

	// Export
	mux.HandleFunc("GET /debates/{id}/export/{format}", h.handleExportDebate)

	// API
	mux.HandleFunc("GET /api/providers", h.handleAPIProviders)
	mux.HandleFunc("GET /api/debates", h.handleAPIDebates)
	mux.HandleFunc("GET /api/debates/{id}", h.handleAPIDebate)
}

// Page handlers

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/debates", http.StatusSeeOther)
}

func (h *Handler) handleDebatesList(w http.ResponseWriter, r *http.Request) {
	debates, err := h.engine.ListDebates(50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Debates":   debates,
		"Providers": h.registry.Available(),
		"Personas":  h.getAllPersonas(),
		"Styles":    h.getAllStyles(),
	}

	h.render(w, "index.html", data)
}

func (h *Handler) handleDebateView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate == nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Debate": debate,
		"Turns":  turns,
	}

	h.render(w, "debate.html", data)
}

func (h *Handler) handleNewDebateForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Providers": h.registry.Available(),
		"Personas":  h.getAllPersonas(),
		"Styles":    h.getAllStyles(),
	}
	h.render(w, "new.html", data)
}

// Partial handlers (for HTMX)

func (h *Handler) handleDebatesPartial(w http.ResponseWriter, r *http.Request) {
	debates, err := h.engine.ListDebates(50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "partials/debates-list.html", map[string]interface{}{
		"Debates": debates,
	})
}

func (h *Handler) handleDebatePartial(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate == nil {
		http.NotFound(w, r)
		return
	}

	h.render(w, "partials/debate-content.html", map[string]interface{}{
		"Debate": debate,
		"Turns":  turns,
	})
}

func (h *Handler) handleDebateTurnsPartial(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate == nil {
		http.NotFound(w, r)
		return
	}

	h.render(w, "partials/turns.html", map[string]interface{}{
		"Debate": debate,
		"Turns":  turns,
	})
}

// Action handlers

func (h *Handler) handleCreateDebate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	maxTurns, _ := strconv.Atoi(r.FormValue("max_turns"))
	if maxTurns <= 0 {
		maxTurns = 5
	}

	// Parse mode
	mode := core.ModeAutomatic
	if r.FormValue("mode") == "turn_by_turn" {
		mode = core.ModeTurnByTurn
	}

	config := core.NewDebateConfig{
		Topic:          r.FormValue("topic"),
		AgentAProvider: r.FormValue("agent_a_provider"),
		AgentAModel:    r.FormValue("agent_a_model"),
		AgentAPersona:  r.FormValue("agent_a_persona"),
		AgentBProvider: r.FormValue("agent_b_provider"),
		AgentBModel:    r.FormValue("agent_b_model"),
		AgentBPersona:  r.FormValue("agent_b_persona"),
		Style:          r.FormValue("style"),
		Mode:           mode,
		MaxTurns:       maxTurns,
	}

	debate, err := h.engine.CreateDebate(r.Context(), config)
	if err != nil {
		h.htmxError(w, err.Error())
		return
	}

	// Check if auto-run is requested
	autoRun := r.FormValue("auto_run") == "on"

	if autoRun && mode == core.ModeAutomatic {
		// Run debate in background
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			h.engine.RunDebate(ctx, debate.ID, nil)
		}()
	}

	// Redirect to debate view
	w.Header().Set("HX-Redirect", "/debates/"+debate.ID)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) handleRunDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Run debate in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		h.engine.RunDebate(ctx, id, nil)
	}()

	w.Header().Set("HX-Trigger", "debateStarted")
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleNextTurn(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	turn, err := h.engine.ExecuteNextTurn(ctx, id)
	if err != nil {
		h.htmxError(w, err.Error())
		return
	}

	debate, turns, _ := h.engine.GetDebateWithTurns(id)
	h.render(w, "partials/turns.html", map[string]interface{}{
		"Debate":  debate,
		"Turns":   turns,
		"NewTurn": turn,
	})
}

func (h *Handler) handleDeleteDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Check if debate is read-only
	debate, err := h.engine.GetDebate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate != nil && debate.ReadOnly {
		h.htmxError(w, "Cannot delete a read-only debate")
		return
	}

	if err := h.engine.DeleteDebate(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/debates")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleExportDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	format := r.PathValue("format")

	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate == nil {
		http.NotFound(w, r)
		return
	}

	exporter, err := export.GetExporter(export.Format(format))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filename := export.GenerateFilename(debate, exporter.FileExtension())

	// Set appropriate content type
	switch format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
	case "json":
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "text/markdown")
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if err := exporter.Export(debate, turns, w); err != nil {
		log.Printf("Export error: %v", err)
		http.Error(w, "Export failed", http.StatusInternalServerError)
	}
}

func (h *Handler) handleLockDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sqlStore, ok := h.storage.(*storage.SQLiteStorage)
	if !ok {
		h.htmxError(w, "Lock not supported for this storage type")
		return
	}

	if err := sqlStore.SetReadOnly(id, true); err != nil {
		h.htmxError(w, err.Error())
		return
	}

	w.Header().Set("HX-Redirect", "/debates/"+id)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleUnlockDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sqlStore, ok := h.storage.(*storage.SQLiteStorage)
	if !ok {
		h.htmxError(w, "Lock not supported for this storage type")
		return
	}

	if err := sqlStore.SetReadOnly(id, false); err != nil {
		h.htmxError(w, err.Error())
		return
	}

	w.Header().Set("HX-Redirect", "/debates/"+id)
	w.WriteHeader(http.StatusOK)
}

// API handlers (JSON)

func (h *Handler) handleAPIProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.registry.List()
	result := make([]map[string]interface{}, 0, len(providers))

	for _, p := range providers {
		result = append(result, map[string]interface{}{
			"name":         p.Name(),
			"display_name": p.DisplayName(),
			"available":    p.Available(),
		})
	}

	h.json(w, result)
}

func (h *Handler) handleAPIDebates(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	debates, err := h.engine.ListDebates(limit, offset)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, debates)
}

func (h *Handler) handleAPIDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate == nil {
		h.jsonError(w, "debate not found", http.StatusNotFound)
		return
	}

	h.json(w, map[string]interface{}{
		"debate": debate,
		"turns":  turns,
	})
}

// Helper methods

func (h *Handler) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (h *Handler) json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) htmxError(w http.ResponseWriter, message string) {
	w.Header().Set("HX-Retarget", "#error-container")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">` + template.HTMLEscapeString(message) + `</div>`))
}

// getAllPersonas returns all personas (builtin + custom from storage).
func (h *Handler) getAllPersonas() []persona.Persona {
	// Start with builtins
	result := persona.DefaultPersonas()

	// Add custom personas from storage
	custom, err := h.storage.ListPersonas(false) // false = exclude builtins (we already have them)
	if err != nil {
		return result
	}

	for _, p := range custom {
		result = append(result, persona.Persona{
			ID:           p.ID,
			Name:         p.Name,
			Description:  p.Description,
			SystemPrompt: p.SystemPrompt,
		})
	}

	return result
}

// getAllStyles returns all styles (builtin + custom from storage).
func (h *Handler) getAllStyles() []style.Style {
	// Start with builtins
	result := style.DefaultStyles()

	// Add custom styles from storage
	custom, err := h.storage.ListStyles(false) // false = exclude builtins (we already have them)
	if err != nil {
		return result
	}

	for _, s := range custom {
		result = append(result, style.Style{
			ID:               s.ID,
			Name:             s.Name,
			Description:      s.Description,
			OpeningPrompt:    s.OpeningPrompt,
			ResponsePrompt:   s.ResponsePrompt,
			ConclusionPrompt: s.ConclusionPrompt,
		})
	}

	return result
}
