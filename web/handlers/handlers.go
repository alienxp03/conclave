// Package handlers provides HTTP handlers for the web interface.
package handlers

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alienxp03/conclave/internal/core"
	"github.com/alienxp03/conclave/internal/council"
	"github.com/alienxp03/conclave/internal/engine"
	"github.com/alienxp03/conclave/internal/export"
	"github.com/alienxp03/conclave/internal/persona"
	"github.com/alienxp03/conclave/internal/provider"
	"github.com/alienxp03/conclave/internal/storage"
	"github.com/alienxp03/conclave/internal/style"
	"github.com/alienxp03/conclave/internal/workspace"
)

//go:embed templates/*.html
var templateFS embed.FS

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	engine        *engine.Engine
	councilEngine *council.Engine
	registry      *provider.Registry
	storage       storage.Storage
	templates     *template.Template
	workspaces    *workspace.Manager
}

// New creates a new Handler.
func New(store storage.Storage, registry *provider.Registry, workspaces *workspace.Manager) *Handler {
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
		slog.Error("Failed to parse templates", "error", err)
		panic(err)
	}

	return &Handler{
		engine:        engine.New(store, registry, workspaces),
		councilEngine: council.New(store, registry),
		registry:      registry,
		storage:       store,
		templates:     tmpl,
		workspaces:    workspaces,
	}
}

// RegisterRoutes registers all HTTP routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// API routes (must be registered first for proper routing)
	mux.HandleFunc("GET /api/providers", h.handleAPIProviders)
	mux.HandleFunc("GET /api/providers/health", h.handleAPIProvidersHealth)
	mux.HandleFunc("GET /api/debates", h.handleAPIDebates)
	mux.HandleFunc("GET /api/debates/{id}", h.handleAPIDebate)
	mux.HandleFunc("GET /api/debates/{id}/stream", h.handleDebateStream)
	mux.HandleFunc("GET /api/projects", h.handleAPIListProjects)
	mux.HandleFunc("POST /api/projects", h.handleAPICreateProject)
	mux.HandleFunc("GET /api/projects/{id}", h.handleAPIGetProject)
	mux.HandleFunc("PUT /api/projects/{id}", h.handleAPIUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", h.handleAPIDeleteProject)

	// Council routes
	mux.HandleFunc("GET /api/councils", h.handleAPIListCouncils)
	mux.HandleFunc("POST /api/councils", h.handleAPICreateCouncil)
	mux.HandleFunc("GET /api/councils/{id}", h.handleAPIGetCouncil)
	mux.HandleFunc("GET /api/councils/{id}/stream", h.handleCouncilStream)

	// New API routes
	mux.HandleFunc("GET /api/personas", h.handleAPIListPersonas)
	mux.HandleFunc("GET /api/styles", h.handleAPIListStyles)
	mux.HandleFunc("GET /api/system/info", h.handleAPISystemInfo)
	mux.HandleFunc("POST /api/debates", h.handleAPICreateDebate)
	mux.HandleFunc("POST /api/debates/{id}/followup", h.handleAPIDebateFollowUp)
	mux.HandleFunc("POST /api/councils/{id}/followup", h.handleAPICouncilFollowUp)
	mux.HandleFunc("DELETE /api/debates/{id}", h.handleAPIDeleteDebate)

	// Actions (POST/DELETE endpoints)
	mux.HandleFunc("POST /debates", h.handleCreateDebate)
	mux.HandleFunc("POST /debates/{id}/run", h.handleRunDebate)
	mux.HandleFunc("POST /debates/{id}/next", h.handleNextTurn)
	mux.HandleFunc("POST /debates/{id}/lock", h.handleLockDebate)
	mux.HandleFunc("POST /debates/{id}/unlock", h.handleUnlockDebate)
	mux.HandleFunc("DELETE /debates/{id}", h.handleDeleteDebate)

	// Export endpoints
	mux.HandleFunc("GET /debates/{id}/export/{format}", h.handleExportDebate)

	// React SPA (catch-all, must be last)
	h.RegisterSPARoutes(mux)
}

// Page handlers

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// Go directly to new debate page
	http.Redirect(w, r, "/new", http.StatusSeeOther)
}

func (h *Handler) handleDebatesList(w http.ResponseWriter, r *http.Request) {
	debates, err := h.engine.ListDebates(50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cwd, _ := os.Getwd()
	data := map[string]interface{}{
		"Debates":   debates,
		"Providers": h.registry.Available(),
		"Personas":  h.getAllPersonas(),
		"Styles":    h.getAllStyles(),
		"CWD":       cwd,
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
	cwd, _ := os.Getwd()
	data := map[string]interface{}{
		"Providers": h.registry.Available(),
		"Personas":  h.getAllPersonas(),
		"Styles":    h.getAllStyles(),
		"CWD":       cwd,
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

	config := core.NewDebateConfig{
		Topic:          r.FormValue("topic"),
		AgentAProvider: r.FormValue("agent_a_provider"),
		AgentAModel:    r.FormValue("agent_a_model"),
		AgentAPersona:  r.FormValue("agent_a_persona"),
		AgentBProvider: r.FormValue("agent_b_provider"),
		AgentBModel:    r.FormValue("agent_b_model"),
		AgentBPersona:  r.FormValue("agent_b_persona"),
		Style:          r.FormValue("style"),
		MaxTurns:       maxTurns,
	}

	debate, err := h.engine.CreateDebate(r.Context(), config)
	if err != nil {
		h.htmxError(w, err.Error())
		return
	}

	// Check if auto-run is requested
	autoRun := r.FormValue("auto_run") == "on"

	if autoRun {
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
		slog.Error("Export failed", "debate_id", id, "format", format, "error", err)
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
		if p.Name() == "mock" {
			continue
		}
		result = append(result, map[string]interface{}{
			"name":          p.Name(),
			"display_name":  p.DisplayName(),
			"available":     p.Available(),
			"models":        p.Models(),
			"default_model": p.DefaultModel(),
		})
	}

	h.json(w, result)
}

func (h *Handler) handleAPIProvidersHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providers := h.registry.List()
	result := make(map[string]interface{})

	for _, p := range providers {
		if p.Name() == "mock" {
			continue
		}

		// Run health check
		status := p.HealthCheck(ctx)

		result[p.Name()] = map[string]interface{}{
			"available":     status.Available,
			"response_time": status.ResponseTime.Seconds(),
			"error":         status.Error,
			"checked_at":    status.CheckedAt,
		}
	}

	h.json(w, map[string]interface{}{
		"providers": result,
	})
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

	// Compute usage stats from turns
	stats := core.ComputeDebateStats(turns, debate.AgentA.ID, debate.AgentB.ID)

	h.json(w, map[string]interface{}{
		"debate": debate,
		"turns":  turns,
		"stats":  stats,
	})
}

func (h *Handler) handleAPIListPersonas(w http.ResponseWriter, r *http.Request) {
	h.json(w, h.getAllPersonas())
}

func (h *Handler) handleAPIListStyles(w http.ResponseWriter, r *http.Request) {
	h.json(w, h.getAllStyles())
}

func (h *Handler) handleAPISystemInfo(w http.ResponseWriter, r *http.Request) {
	cwd, _ := os.Getwd()
	h.json(w, map[string]string{
		"cwd": cwd,
	})
}

func (h *Handler) handleAPIListProjects(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	projects, err := h.storage.ListProjects(limit, offset)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, projects)
}

func (h *Handler) handleAPIGetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := h.storage.GetProject(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.NotFound(w, r)
		return
	}

	debates, err := h.storage.ListDebatesByProject(id, 50, 0)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	councils, err := h.storage.ListCouncilsByProject(id, 50, 0)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"project":  project,
		"debates":  debates,
		"councils": councils,
	})
}

func (h *Handler) handleAPICreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Instructions string `json:"instructions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		h.jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	project := &core.Project{
		ID:           core.GenerateID(),
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		Instructions: req.Instructions,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.storage.CreateProject(project); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, project)
}

func (h *Handler) handleAPIUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := h.storage.GetProject(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.NotFound(w, r)
		return
	}

	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Instructions string `json:"instructions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	project.Name = strings.TrimSpace(req.Name)
	project.Description = strings.TrimSpace(req.Description)
	project.Instructions = req.Instructions
	project.UpdatedAt = time.Now()

	if err := h.storage.UpdateProject(project); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, project)
}

func (h *Handler) handleAPIDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := h.storage.GetProject(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.NotFound(w, r)
		return
	}

	if err := h.storage.DeleteProject(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleAPICreateDebate(w http.ResponseWriter, r *http.Request) {
	type CreateRequest struct {
		core.NewDebateConfig
		AutoRun bool `json:"auto_run"`
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Topic == "" {
		h.jsonError(w, "topic is required", http.StatusBadRequest)
		return
	}

	debate, err := h.engine.CreateDebate(r.Context(), req.NewDebateConfig)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.AutoRun {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			h.engine.RunDebate(ctx, debate.ID, nil)
		}()
	}

	h.json(w, debate)
}

func (h *Handler) handleAPIDeleteDebate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	debate, err := h.engine.GetDebate(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if debate != nil && debate.ReadOnly {
		h.jsonError(w, "Cannot delete a read-only debate", http.StatusForbidden)
		return
	}

	if err := h.engine.DeleteDebate(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleAPIDebateFollowUp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		h.jsonError(w, "content is required", http.StatusBadRequest)
		return
	}

	if err := h.engine.AddFollowUp(r.Context(), id, req.Content); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleAPICouncilFollowUp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		h.jsonError(w, "content is required", http.StatusBadRequest)
		return
	}

	if err := h.councilEngine.AddFollowUp(r.Context(), id, req.Content); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Council Handlers

func (h *Handler) handleAPIListCouncils(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	councils, err := h.storage.ListCouncils(limit, offset)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, councils)
}

func (h *Handler) handleAPIGetCouncil(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	council, err := h.storage.GetCouncil(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses, err := h.storage.GetResponses(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rankings, err := h.storage.GetRankings(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Replace masked names with real names for display
	h.councilEngine.ReplaceMaskedNamesInCouncilData(council, responses, rankings, council.Syntheses)

	h.json(w, map[string]interface{}{
		"council":   council,
		"responses": responses,
		"rankings":  rankings,
	})
}

func (h *Handler) handleAPICreateCouncil(w http.ResponseWriter, r *http.Request) {
	type CreateRequest struct {
		core.NewCouncilConfig
		AutoRun bool `json:"auto_run"`
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Topic == "" {
		h.jsonError(w, "topic is required", http.StatusBadRequest)
		return
	}
	if len(req.Members) < 2 {
		h.jsonError(w, "at least 2 members are required", http.StatusBadRequest)
		return
	}

	c, err := h.councilEngine.CreateCouncil(r.Context(), req.NewCouncilConfig)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.AutoRun {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			h.councilEngine.RunCouncil(ctx, c)
		}()
	}

	h.json(w, c)
}

func (h *Handler) handleCouncilStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slog.Debug("New council stream connection", "id", id, "remote_addr", r.RemoteAddr)

	// Prepare SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Error("Streaming unsupported: ResponseWriter does not implement http.Flusher")
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Helper to send events
	sendEvent := func(event string, data interface{}) {
		jsonData, _ := json.Marshal(data)
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData); err != nil {
			slog.Error("Failed to write SSE event", "event", event, "error", err)
			return
		}
		flusher.Flush()
		slog.Debug("Sent SSE event", "event", event, "data_len", len(jsonData))
	}

	// Get council to make sure it exists
	c, err := h.storage.GetCouncil(id)
	if err != nil {
		slog.Error("Council not found for stream", "id", id, "error", err)
		sendEvent("error", map[string]string{"message": err.Error()})
		return
	}

	// Replace masked names with real names for display
	responses, _ := h.storage.GetResponses(id)
	rankings, _ := h.storage.GetRankings(id)
	h.councilEngine.ReplaceMaskedNamesInCouncilData(c, responses, rankings, c.Syntheses)

	// Send current state immediately
	for _, r := range responses {
		sendEvent("response_collected", map[string]interface{}{
			"id":         r.ID,
			"council_id": r.CouncilID,
			"member_id":  r.MemberID,
			"agent_id":   r.MemberID, // for backward compat with UI
			"round":      r.Round,
			"content":    r.Content,
			"created_at": r.CreatedAt,
		})
	}

	for _, r := range rankings {
		sendEvent("ranking_collected", map[string]interface{}{
			"id":          r.ID,
			"council_id":  r.CouncilID,
			"reviewer_id": r.ReviewerID,
			"agent_id":    r.ReviewerID, // for backward compat with UI
			"round":       r.Round,
			"content":     r.Reasoning,
			"reasoning":   r.Reasoning,
			"created_at":  r.CreatedAt,
		})
	}

	for _, s := range c.Syntheses {
		sendEvent("synthesis_complete", map[string]interface{}{
			"round":     s.Round,
			"synthesis": s.Content,
		})
	}

	// Determine completed stages from existing data
	if len(responses) == len(c.Members) {
		sendEvent("stage_complete", map[string]interface{}{"stage": 1})
	}
	if len(rankings) == len(c.Members) {
		sendEvent("stage_complete", map[string]interface{}{"stage": 2})
	}

	if c.Status == core.StatusCompleted {
		slog.Debug("Council already completed", "id", id)
		sendEvent("complete", c)
		return
	}

	// If it's already running (or failed/pending and we want to restart/continue),
	// we assume the client is connecting to watch progress.
	// NOTE: In a real system, we'd tap into the running process.
	// Here, for simplicity, if it's pending, we start it.
	// If it's in_progress, we might need a way to attach.
	// For now, let's assume this endpoint triggers the run if pending.

	if c.Status == core.StatusPending || c.Status == core.StatusFailed {
		slog.Info("Starting council execution via stream", "id", id)
		callbacks := &council.CouncilCallbacks{
			OnResponseCollected: func(resp core.Response) {
				slog.Debug("Response collected callback", "agent", resp.MemberID)
				sendEvent("response_collected", map[string]interface{}{
					"id":         resp.ID,
					"council_id": resp.CouncilID,
					"member_id":  resp.MemberID,
					"agent_id":   resp.MemberID,
					"round":      resp.Round,
					"content":    resp.Content,
					"created_at": resp.CreatedAt,
				})
			},
			OnRankingCollected: func(ranking core.Ranking) {
				slog.Debug("Ranking collected callback", "agent", ranking.ReviewerID)
				sendEvent("ranking_collected", map[string]interface{}{
					"id":          ranking.ID,
					"council_id":  ranking.CouncilID,
					"reviewer_id": ranking.ReviewerID,
					"agent_id":    ranking.ReviewerID,
					"round":       ranking.Round,
					"content":     ranking.Reasoning,
					"reasoning":   ranking.Reasoning,
					"created_at":  ranking.CreatedAt,
				})
			},
			OnSynthesisComplete: func(synthesis core.CouncilSynthesis) {
				slog.Debug("Synthesis complete callback")
				sendEvent("synthesis_complete", map[string]interface{}{
					"round":     synthesis.Round,
					"synthesis": synthesis.Content,
				})
			},
			OnStageComplete: func(stage int) {
				slog.Debug("Stage complete callback", "stage", stage)
				sendEvent("stage_complete", map[string]interface{}{
					"stage": stage,
				})
			},
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
		defer cancel()

		// Monitor context cancellation
		go func() {
			<-ctx.Done()
			slog.Debug("Stream context done", "id", id, "error", ctx.Err())
		}()

		err := h.councilEngine.RunCouncilWithCallbacks(ctx, c, callbacks)
		if err != nil {
			slog.Error("Council execution failed", "id", id, "error", err)
			sendEvent("error", map[string]string{"message": err.Error()})
			return
		}

		slog.Info("Council execution completed successfully", "id", id)
		sendEvent("complete", c)
	} else if c.Status == core.StatusInProgress {
		// If in progress, poll for updates so late-joining clients get them
		slog.Info("Attaching to in-progress council via polling", "id", id)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		lastResponseCount := len(responses)
		lastRankingCount := len(rankings)
		lastSynthesisCount := len(c.Syntheses)

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				currCouncil, _ := h.storage.GetCouncil(id)
				currResponses, _ := h.storage.GetResponses(id)
				currRankings, _ := h.storage.GetRankings(id)

				// Send new responses
				if len(currResponses) > lastResponseCount {
					for i := lastResponseCount; i < len(currResponses); i++ {
						r := currResponses[i]
						sendEvent("response_collected", map[string]interface{}{
							"id":         r.ID,
							"council_id": r.CouncilID,
							"member_id":  r.MemberID,
							"agent_id":   r.MemberID,
							"round":      r.Round,
							"content":    r.Content,
							"created_at": r.CreatedAt,
						})
					}
					lastResponseCount = len(currResponses)
					if lastResponseCount == len(c.Members) {
						sendEvent("stage_complete", map[string]interface{}{"stage": 1})
					}
				}

				// Send new rankings
				if len(currRankings) > lastRankingCount {
					for i := lastRankingCount; i < len(currRankings); i++ {
						r := currRankings[i]
						sendEvent("ranking_collected", map[string]interface{}{
							"id":          r.ID,
							"council_id":  r.CouncilID,
							"reviewer_id": r.ReviewerID,
							"agent_id":    r.ReviewerID,
							"round":       r.Round,
							"content":     r.Reasoning,
							"reasoning":   r.Reasoning,
							"created_at":  r.CreatedAt,
						})
					}
					lastRankingCount = len(currRankings)
					if lastRankingCount == len(c.Members) {
						sendEvent("stage_complete", map[string]interface{}{"stage": 2})
					}
				}

				// Send synthesis
				if len(currCouncil.Syntheses) > lastSynthesisCount {
					for i := lastSynthesisCount; i < len(currCouncil.Syntheses); i++ {
						s := currCouncil.Syntheses[i]
						sendEvent("synthesis_complete", map[string]interface{}{
							"round":     s.Round,
							"synthesis": s.Content,
						})
					}
					lastSynthesisCount = len(currCouncil.Syntheses)
				}

				if currCouncil.Status == core.StatusCompleted {
					sendEvent("complete", currCouncil)
					return
				}
				if currCouncil.Status == core.StatusFailed {
					sendEvent("error", map[string]string{"message": "Council failed"})
					return
				}
			}
		}
	}
}

// Helper methods

func (h *Handler) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("Template error", "template", name, "error", err)
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
	w.Write([]byte(`<div class="animate-fadeIn bg-red-900 bg-opacity-50 border-2 border-red-500 text-red-200 px-6 py-4 rounded-xl mb-4 shadow-lg">
		<div class="flex items-start">
			<svg class="w-6 h-6 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
			</svg>
			<div>
				<h3 class="font-semibold mb-1">Error</h3>
				<p class="text-sm">` + template.HTMLEscapeString(message) + `</p>
			</div>
		</div>
	</div>`))
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
