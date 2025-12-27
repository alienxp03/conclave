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

	CREATE TABLE IF NOT EXISTS personas (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		system_prompt TEXT NOT NULL,
		is_builtin INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS styles (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		opening_prompt TEXT NOT NULL,
		response_prompt TEXT NOT NULL,
		conclusion_prompt TEXT NOT NULL,
		is_builtin INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_turns_debate_id ON turns(debate_id);
	CREATE INDEX IF NOT EXISTS idx_debates_status ON debates(status);
	CREATE INDEX IF NOT EXISTS idx_debates_created_at ON debates(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_personas_is_builtin ON personas(is_builtin);
	CREATE INDEX IF NOT EXISTS idx_styles_is_builtin ON styles(is_builtin);
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

// Persona represents a stored persona.
type Persona struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	SystemPrompt string    `json:"system_prompt"`
	IsBuiltin    bool      `json:"is_builtin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Style represents a stored debate style.
type Style struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	OpeningPrompt    string    `json:"opening_prompt"`
	ResponsePrompt   string    `json:"response_prompt"`
	ConclusionPrompt string    `json:"conclusion_prompt"`
	IsBuiltin        bool      `json:"is_builtin"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreatePersona creates a new persona.
func (s *SQLiteStorage) CreatePersona(p *Persona) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	query := `
	INSERT INTO personas (id, name, description, system_prompt, is_builtin, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	isBuiltin := 0
	if p.IsBuiltin {
		isBuiltin = 1
	}

	_, err := s.db.Exec(query, p.ID, p.Name, p.Description, p.SystemPrompt, isBuiltin, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create persona: %w", err)
	}
	return nil
}

// GetPersona retrieves a persona by ID.
func (s *SQLiteStorage) GetPersona(id string) (*Persona, error) {
	query := `
	SELECT id, name, description, system_prompt, is_builtin, created_at, updated_at
	FROM personas
	WHERE id = ?
	`

	var p Persona
	var isBuiltin int
	err := s.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.SystemPrompt, &isBuiltin, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}

	p.IsBuiltin = isBuiltin == 1
	return &p, nil
}

// UpdatePersona updates an existing persona.
func (s *SQLiteStorage) UpdatePersona(p *Persona) error {
	// Don't allow updating builtin personas
	existing, err := s.GetPersona(p.ID)
	if err != nil {
		return err
	}
	if existing != nil && existing.IsBuiltin {
		return fmt.Errorf("cannot update builtin persona")
	}

	p.UpdatedAt = time.Now()

	query := `
	UPDATE personas
	SET name = ?, description = ?, system_prompt = ?, updated_at = ?
	WHERE id = ? AND is_builtin = 0
	`

	_, err = s.db.Exec(query, p.Name, p.Description, p.SystemPrompt, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update persona: %w", err)
	}
	return nil
}

// DeletePersona deletes a persona.
func (s *SQLiteStorage) DeletePersona(id string) error {
	// Don't allow deleting builtin personas
	existing, err := s.GetPersona(id)
	if err != nil {
		return err
	}
	if existing != nil && existing.IsBuiltin {
		return fmt.Errorf("cannot delete builtin persona")
	}

	_, err = s.db.Exec("DELETE FROM personas WHERE id = ? AND is_builtin = 0", id)
	if err != nil {
		return fmt.Errorf("failed to delete persona: %w", err)
	}
	return nil
}

// ListPersonas returns all personas.
func (s *SQLiteStorage) ListPersonas(includeBuiltin bool) ([]*Persona, error) {
	query := `
	SELECT id, name, description, system_prompt, is_builtin, created_at, updated_at
	FROM personas
	`
	if !includeBuiltin {
		query += " WHERE is_builtin = 0"
	}
	query += " ORDER BY is_builtin DESC, name ASC"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer rows.Close()

	var personas []*Persona
	for rows.Next() {
		var p Persona
		var isBuiltin int
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.SystemPrompt, &isBuiltin, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		p.IsBuiltin = isBuiltin == 1
		personas = append(personas, &p)
	}

	return personas, nil
}

// CreateStyle creates a new style.
func (s *SQLiteStorage) CreateStyle(st *Style) error {
	now := time.Now()
	st.CreatedAt = now
	st.UpdatedAt = now

	query := `
	INSERT INTO styles (id, name, description, opening_prompt, response_prompt, conclusion_prompt, is_builtin, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isBuiltin := 0
	if st.IsBuiltin {
		isBuiltin = 1
	}

	_, err := s.db.Exec(query, st.ID, st.Name, st.Description, st.OpeningPrompt, st.ResponsePrompt, st.ConclusionPrompt, isBuiltin, st.CreatedAt, st.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create style: %w", err)
	}
	return nil
}

// GetStyle retrieves a style by ID.
func (s *SQLiteStorage) GetStyle(id string) (*Style, error) {
	query := `
	SELECT id, name, description, opening_prompt, response_prompt, conclusion_prompt, is_builtin, created_at, updated_at
	FROM styles
	WHERE id = ?
	`

	var st Style
	var isBuiltin int
	err := s.db.QueryRow(query, id).Scan(
		&st.ID, &st.Name, &st.Description, &st.OpeningPrompt, &st.ResponsePrompt, &st.ConclusionPrompt, &isBuiltin, &st.CreatedAt, &st.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get style: %w", err)
	}

	st.IsBuiltin = isBuiltin == 1
	return &st, nil
}

// UpdateStyle updates an existing style.
func (s *SQLiteStorage) UpdateStyle(st *Style) error {
	existing, err := s.GetStyle(st.ID)
	if err != nil {
		return err
	}
	if existing != nil && existing.IsBuiltin {
		return fmt.Errorf("cannot update builtin style")
	}

	st.UpdatedAt = time.Now()

	query := `
	UPDATE styles
	SET name = ?, description = ?, opening_prompt = ?, response_prompt = ?, conclusion_prompt = ?, updated_at = ?
	WHERE id = ? AND is_builtin = 0
	`

	_, err = s.db.Exec(query, st.Name, st.Description, st.OpeningPrompt, st.ResponsePrompt, st.ConclusionPrompt, st.UpdatedAt, st.ID)
	if err != nil {
		return fmt.Errorf("failed to update style: %w", err)
	}
	return nil
}

// DeleteStyle deletes a style.
func (s *SQLiteStorage) DeleteStyle(id string) error {
	existing, err := s.GetStyle(id)
	if err != nil {
		return err
	}
	if existing != nil && existing.IsBuiltin {
		return fmt.Errorf("cannot delete builtin style")
	}

	_, err = s.db.Exec("DELETE FROM styles WHERE id = ? AND is_builtin = 0", id)
	if err != nil {
		return fmt.Errorf("failed to delete style: %w", err)
	}
	return nil
}

// ListStyles returns all styles.
func (s *SQLiteStorage) ListStyles(includeBuiltin bool) ([]*Style, error) {
	query := `
	SELECT id, name, description, opening_prompt, response_prompt, conclusion_prompt, is_builtin, created_at, updated_at
	FROM styles
	`
	if !includeBuiltin {
		query += " WHERE is_builtin = 0"
	}
	query += " ORDER BY is_builtin DESC, name ASC"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list styles: %w", err)
	}
	defer rows.Close()

	var styles []*Style
	for rows.Next() {
		var st Style
		var isBuiltin int
		err := rows.Scan(&st.ID, &st.Name, &st.Description, &st.OpeningPrompt, &st.ResponsePrompt, &st.ConclusionPrompt, &isBuiltin, &st.CreatedAt, &st.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan style: %w", err)
		}
		st.IsBuiltin = isBuiltin == 1
		styles = append(styles, &st)
	}

	return styles, nil
}
