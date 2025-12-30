// Package storage provides persistence for debate sessions.
package storage

import (
	"github.com/alienxp03/dbate/internal/core"
)

// Storage defines the interface for debate persistence.
type Storage interface {
	// Initialize sets up the storage (creates tables, etc.)
	Initialize() error

	// Close closes the storage connection.
	Close() error

	// Debate operations
	CreateDebate(debate *core.Debate) error
	GetDebate(id string) (*core.Debate, error)
	UpdateDebate(debate *core.Debate) error
	DeleteDebate(id string) error
	ListDebates(limit, offset int) ([]*core.DebateSummary, error)

	// Turn operations
	AddTurn(turn *core.Turn) error
	GetTurns(debateID string) ([]*core.Turn, error)
	GetLatestTurn(debateID string) (*core.Turn, error)

	// Persona operations
	GetPersona(id string) (*Persona, error)
	ListPersonas(includeBuiltin bool) ([]*Persona, error)

	// Style operations
	GetStyle(id string) (*Style, error)
	ListStyles(includeBuiltin bool) ([]*Style, error)
}
