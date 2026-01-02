package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alienxp03/conclave/internal/core"
)

// StreamEvent represents a server-sent event.
type StreamEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// handleDebateStream streams debate updates using Server-Sent Events.
func (h *Handler) handleDebateStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slog.Debug("New debate stream connection", "id", id, "remote_addr", r.RemoteAddr)

	// Set headers for SSE
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

	// Get initial debate state
	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		slog.Error("Failed to get debate for stream", "id", id, "error", err)
		h.sendSSEError(w, flusher, "Failed to get debate")
		return
	}
	if debate == nil {
		slog.Warn("Debate not found for stream", "id", id)
		h.sendSSEError(w, flusher, "Debate not found")
		return
	}

	// Send existing turns immediately
	for _, turn := range turns {
		h.sendSSEEvent(w, flusher, "turn_complete", turn)
	}

	// If debate is not in progress, send complete event and close
	if debate.Status != core.StatusInProgress {
		slog.Debug("Debate already completed", "id", id)
		h.sendSSEEvent(w, flusher, "debate_complete", debate)
		return
	}

	// Poll for updates (in a production app, you'd use a pub/sub system)
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastTurnCount := len(turns)

	for {
		select {
		case <-ctx.Done():
			slog.Debug("Stream context done", "id", id)
			return
		case <-ticker.C:
			// Get updated debate state
			updatedDebate, updatedTurns, err := h.engine.GetDebateWithTurns(id)
			if err != nil {
				slog.Error("Stream error updating debate", "id", id, "error", err)
				continue
			}

			// Check if new turns were added
			if len(updatedTurns) > lastTurnCount {
				slog.Debug("New turns detected", "id", id, "count", len(updatedTurns)-lastTurnCount)
				// Send new turns
				for i := lastTurnCount; i < len(updatedTurns); i++ {
					h.sendSSEEvent(w, flusher, "turn_complete", updatedTurns[i])
				}
				lastTurnCount = len(updatedTurns)
			}

			// Check if debate completed
			if updatedDebate.Status != core.StatusInProgress {
				slog.Debug("Debate completed during stream", "id", id)
				h.sendSSEEvent(w, flusher, "debate_complete", updatedDebate)
				return
			}
		}
	}
}

// sendSSEEvent sends a server-sent event.
func (h *Handler) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to marshal SSE data", "error", err)
		return
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", eventType); err != nil {
		slog.Error("Failed to write SSE event", "error", err)
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
		slog.Error("Failed to write SSE data", "error", err)
		return
	}
	flusher.Flush()
}

// sendSSEError sends an error event.
func (h *Handler) sendSSEError(w http.ResponseWriter, flusher http.Flusher, message string) {
	errorData := map[string]string{"message": message}
	h.sendSSEEvent(w, flusher, "error", errorData)
}