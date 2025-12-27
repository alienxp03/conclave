package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alienxp03/dbate/internal/core"
)

// SQLiteStorage implements Storage using SQLite.
type SQLiteStorage struct {
	db   *sql.DB
	path string
}

// NewSQLiteStorage creates a new SQLite storage instance.
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &SQLiteStorage{
		db:   db,
		path: dbPath,
	}, nil
}

// Initialize creates the database schema.
func (s *SQLiteStorage) Initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS debates (
		id TEXT PRIMARY KEY,
		topic TEXT NOT NULL,
		agent_a_json TEXT NOT NULL,
		agent_b_json TEXT NOT NULL,
		style TEXT NOT NULL,
		mode TEXT NOT NULL DEFAULT 'automatic',
		max_turns INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		read_only INTEGER NOT NULL DEFAULT 0,
		conclusion_json TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		completed_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS turns (
		id TEXT PRIMARY KEY,
		debate_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		number INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (debate_id) REFERENCES debates(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_turns_debate_id ON turns(debate_id);
	CREATE INDEX IF NOT EXISTS idx_debates_status ON debates(status);
	CREATE INDEX IF NOT EXISTS idx_debates_created_at ON debates(created_at DESC);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Run migrations for existing databases
	s.migrate()

	return nil
}

// migrate handles schema migrations for existing databases.
func (s *SQLiteStorage) migrate() {
	// Add mode column if not exists
	s.db.Exec("ALTER TABLE debates ADD COLUMN mode TEXT NOT NULL DEFAULT 'automatic'")
	// Add read_only column if not exists
	s.db.Exec("ALTER TABLE debates ADD COLUMN read_only INTEGER NOT NULL DEFAULT 0")
}

// Close closes the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// CreateDebate creates a new debate.
func (s *SQLiteStorage) CreateDebate(debate *core.Debate) error {
	agentAJSON, err := json.Marshal(debate.AgentA)
	if err != nil {
		return fmt.Errorf("failed to marshal agent A: %w", err)
	}

	agentBJSON, err := json.Marshal(debate.AgentB)
	if err != nil {
		return fmt.Errorf("failed to marshal agent B: %w", err)
	}

	var conclusionJSON *string
	if debate.Conclusion != nil {
		data, err := json.Marshal(debate.Conclusion)
		if err != nil {
			return fmt.Errorf("failed to marshal conclusion: %w", err)
		}
		str := string(data)
		conclusionJSON = &str
	}

	mode := debate.Mode
	if mode == "" {
		mode = core.ModeAutomatic
	}

	query := `
	INSERT INTO debates (id, topic, agent_a_json, agent_b_json, style, mode, max_turns, status, read_only, conclusion_json, created_at, updated_at, completed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	readOnly := 0
	if debate.ReadOnly {
		readOnly = 1
	}

	_, err = s.db.Exec(query,
		debate.ID,
		debate.Topic,
		string(agentAJSON),
		string(agentBJSON),
		debate.Style,
		mode,
		debate.MaxTurns,
		debate.Status,
		readOnly,
		conclusionJSON,
		debate.CreatedAt,
		debate.UpdatedAt,
		debate.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert debate: %w", err)
	}

	return nil
}

// GetDebate retrieves a debate by ID.
func (s *SQLiteStorage) GetDebate(id string) (*core.Debate, error) {
	query := `
	SELECT id, topic, agent_a_json, agent_b_json, style, mode, max_turns, status, read_only, conclusion_json, created_at, updated_at, completed_at
	FROM debates
	WHERE id = ?
	`

	var debate core.Debate
	var agentAJSON, agentBJSON string
	var conclusionJSON sql.NullString
	var completedAt sql.NullTime
	var mode sql.NullString
	var readOnly int

	err := s.db.QueryRow(query, id).Scan(
		&debate.ID,
		&debate.Topic,
		&agentAJSON,
		&agentBJSON,
		&debate.Style,
		&mode,
		&debate.MaxTurns,
		&debate.Status,
		&readOnly,
		&conclusionJSON,
		&debate.CreatedAt,
		&debate.UpdatedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get debate: %w", err)
	}

	if err := json.Unmarshal([]byte(agentAJSON), &debate.AgentA); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent A: %w", err)
	}

	if err := json.Unmarshal([]byte(agentBJSON), &debate.AgentB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent B: %w", err)
	}

	if conclusionJSON.Valid {
		var conclusion core.Conclusion
		if err := json.Unmarshal([]byte(conclusionJSON.String), &conclusion); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conclusion: %w", err)
		}
		debate.Conclusion = &conclusion
	}

	if completedAt.Valid {
		debate.CompletedAt = &completedAt.Time
	}

	if mode.Valid {
		debate.Mode = core.DebateMode(mode.String)
	} else {
		debate.Mode = core.ModeAutomatic
	}

	debate.ReadOnly = readOnly == 1

	return &debate, nil
}

// UpdateDebate updates an existing debate.
func (s *SQLiteStorage) UpdateDebate(debate *core.Debate) error {
	agentAJSON, err := json.Marshal(debate.AgentA)
	if err != nil {
		return fmt.Errorf("failed to marshal agent A: %w", err)
	}

	agentBJSON, err := json.Marshal(debate.AgentB)
	if err != nil {
		return fmt.Errorf("failed to marshal agent B: %w", err)
	}

	var conclusionJSON *string
	if debate.Conclusion != nil {
		data, err := json.Marshal(debate.Conclusion)
		if err != nil {
			return fmt.Errorf("failed to marshal conclusion: %w", err)
		}
		str := string(data)
		conclusionJSON = &str
	}

	debate.UpdatedAt = time.Now()

	readOnly := 0
	if debate.ReadOnly {
		readOnly = 1
	}

	query := `
	UPDATE debates
	SET topic = ?, agent_a_json = ?, agent_b_json = ?, style = ?, mode = ?, max_turns = ?, status = ?, read_only = ?, conclusion_json = ?, updated_at = ?, completed_at = ?
	WHERE id = ?
	`

	_, err = s.db.Exec(query,
		debate.Topic,
		string(agentAJSON),
		string(agentBJSON),
		debate.Style,
		debate.Mode,
		debate.MaxTurns,
		debate.Status,
		readOnly,
		conclusionJSON,
		debate.UpdatedAt,
		debate.CompletedAt,
		debate.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update debate: %w", err)
	}

	return nil
}

// DeleteDebate deletes a debate and its turns.
func (s *SQLiteStorage) DeleteDebate(id string) error {
	// Check if debate is read-only
	debate, err := s.GetDebate(id)
	if err != nil {
		return err
	}
	if debate != nil && debate.ReadOnly {
		return fmt.Errorf("cannot delete read-only debate")
	}

	_, err = s.db.Exec("DELETE FROM debates WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete debate: %w", err)
	}
	return nil
}

// ListDebates returns a list of debate summaries.
func (s *SQLiteStorage) ListDebates(limit, offset int) ([]*core.DebateSummary, error) {
	query := `
	SELECT d.id, d.topic, d.status, d.style, d.mode, d.read_only, d.agent_a_json, d.agent_b_json, d.created_at,
		   (SELECT COUNT(*) FROM turns WHERE debate_id = d.id) as turn_count
	FROM debates d
	ORDER BY d.created_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list debates: %w", err)
	}
	defer rows.Close()

	var summaries []*core.DebateSummary
	for rows.Next() {
		var summary core.DebateSummary
		var agentAJSON, agentBJSON string
		var mode sql.NullString
		var readOnly int

		err := rows.Scan(
			&summary.ID,
			&summary.Topic,
			&summary.Status,
			&summary.Style,
			&mode,
			&readOnly,
			&agentAJSON,
			&agentBJSON,
			&summary.CreatedAt,
			&summary.TurnCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan debate summary: %w", err)
		}

		var agentA, agentB core.Agent
		json.Unmarshal([]byte(agentAJSON), &agentA)
		json.Unmarshal([]byte(agentBJSON), &agentB)

		summary.AgentA = fmt.Sprintf("%s:%s", agentA.Provider, agentA.Persona)
		summary.AgentB = fmt.Sprintf("%s:%s", agentB.Provider, agentB.Persona)

		if mode.Valid {
			summary.Mode = core.DebateMode(mode.String)
		} else {
			summary.Mode = core.ModeAutomatic
		}

		summary.ReadOnly = readOnly == 1

		summaries = append(summaries, &summary)
	}

	return summaries, nil
}

// SetReadOnly sets the read-only flag for a debate.
func (s *SQLiteStorage) SetReadOnly(id string, readOnly bool) error {
	val := 0
	if readOnly {
		val = 1
	}

	_, err := s.db.Exec("UPDATE debates SET read_only = ?, updated_at = ? WHERE id = ?", val, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set read-only: %w", err)
	}
	return nil
}

// AddTurn adds a turn to a debate.
func (s *SQLiteStorage) AddTurn(turn *core.Turn) error {
	query := `
	INSERT INTO turns (id, debate_id, agent_id, number, content, created_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		turn.ID,
		turn.DebateID,
		turn.AgentID,
		turn.Number,
		turn.Content,
		turn.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert turn: %w", err)
	}

	return nil
}

// GetTurns returns all turns for a debate.
func (s *SQLiteStorage) GetTurns(debateID string) ([]*core.Turn, error) {
	query := `
	SELECT id, debate_id, agent_id, number, content, created_at
	FROM turns
	WHERE debate_id = ?
	ORDER BY number ASC
	`

	rows, err := s.db.Query(query, debateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get turns: %w", err)
	}
	defer rows.Close()

	var turns []*core.Turn
	for rows.Next() {
		var turn core.Turn
		err := rows.Scan(
			&turn.ID,
			&turn.DebateID,
			&turn.AgentID,
			&turn.Number,
			&turn.Content,
			&turn.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan turn: %w", err)
		}
		turns = append(turns, &turn)
	}

	return turns, nil
}

// GetLatestTurn returns the most recent turn for a debate.
func (s *SQLiteStorage) GetLatestTurn(debateID string) (*core.Turn, error) {
	query := `
	SELECT id, debate_id, agent_id, number, content, created_at
	FROM turns
	WHERE debate_id = ?
	ORDER BY number DESC
	LIMIT 1
	`

	var turn core.Turn
	err := s.db.QueryRow(query, debateID).Scan(
		&turn.ID,
		&turn.DebateID,
		&turn.AgentID,
		&turn.Number,
		&turn.Content,
		&turn.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest turn: %w", err)
	}

	return &turn, nil
}

// DefaultDBPath returns the default database path.
func DefaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "dbate.db"
	}
	return filepath.Join(home, ".dbate", "dbate.db")
}
