package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alienxp03/dbate/internal/core"
)

// StreamEvent represents a server-sent event.
type StreamEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// handleDebateStream streams debate updates using Server-Sent Events.
func (h *Handler) handleDebateStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Get initial debate state
	debate, turns, err := h.engine.GetDebateWithTurns(id)
	if err != nil {
		h.sendSSEError(w, flusher, "Failed to get debate")
		return
	}
	if debate == nil {
		h.sendSSEError(w, flusher, "Debate not found")
		return
	}

	// Send existing turns immediately
	for _, turn := range turns {
		h.sendSSEEvent(w, flusher, "turn_complete", turn)
	}

	// If debate is not in progress, send complete event and close
	if debate.Status != core.StatusInProgress {
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
			return
		case <-ticker.C:
			// Get updated debate state
			updatedDebate, updatedTurns, err := h.engine.GetDebateWithTurns(id)
			if err != nil {
				log.Printf("Stream error for debate %s: %v", id, err)
				continue
			}

			// Check if new turns were added
			if len(updatedTurns) > lastTurnCount {
				// Send new turns
				for i := lastTurnCount; i < len(updatedTurns); i++ {
					h.sendSSEEvent(w, flusher, "turn_complete", updatedTurns[i])
				}
				lastTurnCount = len(updatedTurns)
			}

			// Check if debate completed
			if updatedDebate.Status != core.StatusInProgress {
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
		log.Printf("Failed to marshal SSE data: %v", err)
		return
	}

	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

// sendSSEError sends an error event.
func (h *Handler) sendSSEError(w http.ResponseWriter, flusher http.Flusher, message string) {
	errorData := map[string]string{"message": message}
	h.sendSSEEvent(w, flusher, "error", errorData)
}
