package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alienxp03/conclave/internal/core"
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
		title TEXT NOT NULL DEFAULT 'New conversation',
		topic TEXT NOT NULL,
		cwd TEXT NOT NULL DEFAULT '',
		agent_a_json TEXT NOT NULL,
		agent_b_json TEXT NOT NULL,
		style TEXT NOT NULL,
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

	CREATE TABLE IF NOT EXISTS councils (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL DEFAULT 'New conversation',
		topic TEXT NOT NULL,
		cwd TEXT NOT NULL DEFAULT '',
		chairman_json TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		synthesis TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		completed_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS council_members (
		id TEXT PRIMARY KEY,
		council_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		persona TEXT NOT NULL,
		display_name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (council_id) REFERENCES councils(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS responses (
		id TEXT PRIMARY KEY,
		council_id TEXT NOT NULL,
		member_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (council_id) REFERENCES councils(id) ON DELETE CASCADE,
		FOREIGN KEY (member_id) REFERENCES council_members(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS rankings (
		id TEXT PRIMARY KEY,
		council_id TEXT NOT NULL,
		reviewer_id TEXT NOT NULL,
		rankings_json TEXT NOT NULL,
		reasoning TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (council_id) REFERENCES councils(id) ON DELETE CASCADE,
		FOREIGN KEY (reviewer_id) REFERENCES council_members(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_turns_debate_id ON turns(debate_id);
	CREATE INDEX IF NOT EXISTS idx_debates_status ON debates(status);
	CREATE INDEX IF NOT EXISTS idx_debates_created_at ON debates(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_personas_is_builtin ON personas(is_builtin);
	CREATE INDEX IF NOT EXISTS idx_styles_is_builtin ON styles(is_builtin);
	CREATE INDEX IF NOT EXISTS idx_council_members_council_id ON council_members(council_id);
	CREATE INDEX IF NOT EXISTS idx_responses_council_id ON responses(council_id);
	CREATE INDEX IF NOT EXISTS idx_rankings_council_id ON rankings(council_id);
	CREATE INDEX IF NOT EXISTS idx_councils_status ON councils(status);
	CREATE INDEX IF NOT EXISTS idx_councils_created_at ON councils(created_at DESC);
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
	// Add cwd column if not exists
	s.db.Exec("ALTER TABLE debates ADD COLUMN cwd TEXT NOT NULL DEFAULT ''")
	s.db.Exec("ALTER TABLE councils ADD COLUMN cwd TEXT NOT NULL DEFAULT ''")
	// Add title column if not exists
	s.db.Exec("ALTER TABLE debates ADD COLUMN title TEXT NOT NULL DEFAULT 'New conversation'")
	s.db.Exec("ALTER TABLE councils ADD COLUMN title TEXT NOT NULL DEFAULT 'New conversation'")
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

	if debate.Title == "" {
		debate.Title = "New conversation"
	}

	query := `
	INSERT INTO debates (id, title, topic, cwd, agent_a_json, agent_b_json, style, max_turns, status, read_only, conclusion_json, created_at, updated_at, completed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	readOnly := 0
	if debate.ReadOnly {
		readOnly = 1
	}

	_, err = s.db.Exec(query,
		debate.ID,
		debate.Title,
		debate.Topic,
		debate.CWD,
		string(agentAJSON),
		string(agentBJSON),
		debate.Style,
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
	SELECT id, title, topic, cwd, agent_a_json, agent_b_json, style, max_turns, status, read_only, conclusion_json, created_at, updated_at, completed_at
	FROM debates
	WHERE id = ?
	`

	var debate core.Debate
	var agentAJSON, agentBJSON string
	var conclusionJSON sql.NullString
	var completedAt sql.NullTime
	var readOnly int

	err := s.db.QueryRow(query, id).Scan(
		&debate.ID,
		&debate.Title,
		&debate.Topic,
		&debate.CWD,
		&agentAJSON,
		&agentBJSON,
		&debate.Style,
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

	debate.ReadOnly = readOnly == 1

	return &debate, nil
}

// UpdateDebate updates an existing debate.
func (s *SQLiteStorage) UpdateDebate(debate *core.Debate) error {
	// If title is "New conversation", try to preserve existing title from DB
	if debate.Title == "New conversation" {
		var existingTitle string
		err := s.db.QueryRow("SELECT title FROM debates WHERE id = ?", debate.ID).Scan(&existingTitle)
		if err == nil && existingTitle != "New conversation" {
			debate.Title = existingTitle
		}
	}

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
	SET title = ?, topic = ?, cwd = ?, agent_a_json = ?, agent_b_json = ?, style = ?, max_turns = ?, status = ?, read_only = ?, conclusion_json = ?, updated_at = ?, completed_at = ?
	WHERE id = ?
	`

	_, err = s.db.Exec(query,
		debate.Title,
		debate.Topic,
		debate.CWD,
		string(agentAJSON),
		string(agentBJSON),
		debate.Style,
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

// UpdateDebateTitle updates only the title of a debate.
func (s *SQLiteStorage) UpdateDebateTitle(id, title string) error {
	_, err := s.db.Exec("UPDATE debates SET title = ?, updated_at = ? WHERE id = ?", title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update debate title: %w", err)
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
	SELECT d.id, d.title, d.topic, d.cwd, d.status, d.style, d.read_only, d.agent_a_json, d.agent_b_json, d.created_at,
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
		var readOnly int

		err := rows.Scan(
			&summary.ID,
			&summary.Title,
			&summary.Topic,
			&summary.CWD,
			&summary.Status,
			&summary.Style,
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
		return "conclave.db"
	}
	return filepath.Join(home, ".conclave", "conclave.db")
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

// CreateCouncil creates a new council.
func (s *SQLiteStorage) CreateCouncil(council *core.Council) error {
	if council.Title == "" {
		council.Title = "New conversation"
	}

	chairmanJSON, err := json.Marshal(council.Chairman)
	if err != nil {
		return fmt.Errorf("failed to marshal chairman: %w", err)
	}

	query := `
	INSERT INTO councils (id, title, topic, cwd, chairman_json, status, synthesis, created_at, updated_at, completed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var completedAt *time.Time
	if council.CompletedAt != nil {
		completedAt = council.CompletedAt
	}

	_, err = s.db.Exec(query,
		council.ID,
		council.Title,
		council.Topic,
		council.CWD,
		string(chairmanJSON),
		council.Status,
		council.Synthesis,
		council.CreatedAt,
		council.UpdatedAt,
		completedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert council: %w", err)
	}

	// Insert members
	for _, member := range council.Members {
		err := s.insertCouncilMember(council.ID, member)
		if err != nil {
			return fmt.Errorf("failed to insert member: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStorage) insertCouncilMember(councilID string, member core.Agent) error {
	query := `
	INSERT INTO council_members (id, council_id, provider, model, persona, display_name, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		member.ID,
		councilID,
		member.Provider,
		member.Model,
		member.Persona,
		member.Name,
		time.Now(),
	)

	return err
}

// GetCouncil retrieves a council by ID.
func (s *SQLiteStorage) GetCouncil(id string) (*core.Council, error) {
	query := `
	SELECT id, title, topic, cwd, chairman_json, status, synthesis, created_at, updated_at, completed_at
	FROM councils
	WHERE id = ?
	`

	var council core.Council
	var chairmanJSON string
	var synthesis sql.NullString
	var completedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&council.ID,
		&council.Title,
		&council.Topic,
		&council.CWD,
		&chairmanJSON,
		&council.Status,
		&synthesis,
		&council.CreatedAt,
		&council.UpdatedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("council not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get council: %w", err)
	}

	// Unmarshal chairman
	if err := json.Unmarshal([]byte(chairmanJSON), &council.Chairman); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chairman: %w", err)
	}

	if synthesis.Valid {
		council.Synthesis = synthesis.String
	}

	if completedAt.Valid {
		council.CompletedAt = &completedAt.Time
	}

	// Get members
	members, err := s.getCouncilMembers(id)
	if err != nil {
		return nil, err
	}
	council.Members = members

	return &council, nil
}

func (s *SQLiteStorage) getCouncilMembers(councilID string) ([]core.Agent, error) {
	query := `
	SELECT id, provider, model, persona, display_name
	FROM council_members
	WHERE council_id = ?
	ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, councilID)
	if err != nil {
		return nil, fmt.Errorf("failed to get council members: %w", err)
	}
	defer rows.Close()

	var members []core.Agent
	for rows.Next() {
		var member core.Agent
		err := rows.Scan(
			&member.ID,
			&member.Provider,
			&member.Model,
			&member.Persona,
			&member.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// UpdateCouncil updates an existing council.
func (s *SQLiteStorage) UpdateCouncil(council *core.Council) error {
	// If title is "New conversation", try to preserve existing title from DB
	if council.Title == "New conversation" {
		var existingTitle string
		err := s.db.QueryRow("SELECT title FROM councils WHERE id = ?", council.ID).Scan(&existingTitle)
		if err == nil && existingTitle != "New conversation" {
			council.Title = existingTitle
		}
	}

	chairmanJSON, err := json.Marshal(council.Chairman)
	if err != nil {
		return fmt.Errorf("failed to marshal chairman: %w", err)
	}

	query := `
	UPDATE councils
	SET title = ?, topic = ?, cwd = ?, chairman_json = ?, status = ?, synthesis = ?, updated_at = ?, completed_at = ?
	WHERE id = ?
	`

	var completedAt *time.Time
	if council.CompletedAt != nil {
		completedAt = council.CompletedAt
	}

	_, err = s.db.Exec(query,
		council.Title,
		council.Topic,
		council.CWD,
		string(chairmanJSON),
		council.Status,
		council.Synthesis,
		council.UpdatedAt,
		completedAt,
		council.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update council: %w", err)
	}

	return nil
}

// UpdateCouncilTitle updates only the title of a council.
func (s *SQLiteStorage) UpdateCouncilTitle(id, title string) error {
	_, err := s.db.Exec("UPDATE councils SET title = ?, updated_at = ? WHERE id = ?", title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update council title: %w", err)
	}
	return nil
}

// DeleteCouncil deletes a council.
func (s *SQLiteStorage) DeleteCouncil(id string) error {
	_, err := s.db.Exec("DELETE FROM councils WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete council: %w", err)
	}
	return nil
}

// ListCouncils returns a list of councils.
func (s *SQLiteStorage) ListCouncils(limit, offset int) ([]*core.CouncilSummary, error) {
	query := `
	SELECT
		c.id,
		c.title,
		c.topic,
		c.cwd,
		c.status,
		c.created_at,
		COUNT(cm.id) as member_count
	FROM councils c
	LEFT JOIN council_members cm ON c.id = cm.council_id
	GROUP BY c.id
	ORDER BY c.created_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list councils: %w", err)
	}
	defer rows.Close()

	var summaries []*core.CouncilSummary
	for rows.Next() {
		var summary core.CouncilSummary
		err := rows.Scan(
			&summary.ID,
			&summary.Title,
			&summary.Topic,
			&summary.CWD,
			&summary.Status,
			&summary.CreatedAt,
			&summary.MemberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan council summary: %w", err)
		}
		summaries = append(summaries, &summary)
	}

	return summaries, nil
}

// AddResponse adds a response to a council.
func (s *SQLiteStorage) AddResponse(response *core.Response) error {
	query := `
	INSERT INTO responses (id, council_id, member_id, content, created_at)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		response.ID,
		response.CouncilID,
		response.MemberID,
		response.Content,
		response.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert response: %w", err)
	}

	return nil
}

// GetResponses returns all responses for a council.
func (s *SQLiteStorage) GetResponses(councilID string) ([]*core.Response, error) {
	query := `
	SELECT id, council_id, member_id, content, created_at
	FROM responses
	WHERE council_id = ?
	ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, councilID)
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}
	defer rows.Close()

	var responses []*core.Response
	for rows.Next() {
		var response core.Response
		err := rows.Scan(
			&response.ID,
			&response.CouncilID,
			&response.MemberID,
			&response.Content,
			&response.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response: %w", err)
		}
		responses = append(responses, &response)
	}

	return responses, nil
}

// AddRanking adds a ranking to a council.
func (s *SQLiteStorage) AddRanking(ranking *core.Ranking) error {
	rankingsJSON, err := json.Marshal(ranking.Rankings)
	if err != nil {
		return fmt.Errorf("failed to marshal rankings: %w", err)
	}

	query := `
	INSERT INTO rankings (id, council_id, reviewer_id, rankings_json, reasoning, created_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		ranking.ID,
		ranking.CouncilID,
		ranking.ReviewerID,
		string(rankingsJSON),
		ranking.Reasoning,
		ranking.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert ranking: %w", err)
	}

	return nil
}

// GetRankings returns all rankings for a council.
func (s *SQLiteStorage) GetRankings(councilID string) ([]*core.Ranking, error) {
	query := `
	SELECT id, council_id, reviewer_id, rankings_json, reasoning, created_at
	FROM rankings
	WHERE council_id = ?
	ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, councilID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rankings: %w", err)
	}
	defer rows.Close()

	var rankings []*core.Ranking
	for rows.Next() {
		var ranking core.Ranking
		var rankingsJSON string
		err := rows.Scan(
			&ranking.ID,
			&ranking.CouncilID,
			&ranking.ReviewerID,
			&rankingsJSON,
			&ranking.Reasoning,
			&ranking.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ranking: %w", err)
		}

		// Unmarshal rankings
		if err := json.Unmarshal([]byte(rankingsJSON), &ranking.Rankings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rankings: %w", err)
		}

		rankings = append(rankings, &ranking)
	}

	return rankings, nil
}
