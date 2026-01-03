package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/alienxp03/conclave/internal/core"
)

// Workspace represents a project directory where Conclave sessions can run.
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// Manager handles workspace persistence and operations.
type Manager struct {
	mu         sync.RWMutex
	workspaces map[string]*Workspace
	filePath   string
}

// NewManager creates a new workspace manager.
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".conclave", "workspaces.json")

	m := &Manager{
		workspaces: make(map[string]*Workspace),
		filePath:   path,
	}

	if err := m.load(); err != nil {
		return nil, err
	}

	return m, nil
}

// load reads workspaces from disk.
func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var saved []*Workspace
	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}

	m.workspaces = make(map[string]*Workspace)
	for _, w := range saved {
		m.workspaces[w.ID] = w
	}

	return nil
}

// save writes workspaces to disk.
func (m *Manager) save() error {
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	workspaces := make([]*Workspace, 0, len(m.workspaces))
	for _, w := range m.workspaces {
		workspaces = append(workspaces, w)
	}

	data, err := json.MarshalIndent(workspaces, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// List returns all workspaces sorted by name.
func (m *Manager) List() []*Workspace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Workspace, 0, len(m.workspaces))
	for _, w := range m.workspaces {
		list = append(list, w)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

// Get retrieves a workspace by ID.
func (m *Manager) Get(id string) (*Workspace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w, ok := m.workspaces[id]
	if !ok {
		return nil, fmt.Errorf("workspace not found: %s", id)
	}
	return w, nil
}

// Add creates and saves a new workspace.
func (m *Manager) Add(name, path string) (*Workspace, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates
	for _, w := range m.workspaces {
		if w.Path == absPath {
			return nil, fmt.Errorf("workspace already exists for path: %s", absPath)
		}
	}

	id := core.GenerateID()
	w := &Workspace{
		ID:        id,
		Name:      name,
		Path:      absPath,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	m.workspaces[id] = w

	if err := m.save(); err != nil {
		delete(m.workspaces, id)
		return nil, err
	}

	return w, nil
}

// Delete removes a workspace.
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.workspaces[id]; !ok {
		return fmt.Errorf("workspace not found: %s", id)
	}

	delete(m.workspaces, id)
	return m.save()
}

// GetCurrent returns a workspace representing the current working directory.
func (m *Manager) GetCurrent() *Workspace {
	cwd, _ := os.Getwd()
	name := filepath.Base(cwd)

	// Check if this path is already a saved workspace
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, w := range m.workspaces {
		if w.Path == cwd {
			return w
		}
	}

	// Return ephemeral workspace
	return &Workspace{
		ID:   "cwd",
		Name: fmt.Sprintf("%s (Current)", name),
		Path: cwd,
	}
}

// ResolvePath returns the absolute path for a workspace ID (or handles "cwd" special case).
func (m *Manager) ResolvePath(id string) (string, error) {
	if id == "" || id == "cwd" {
		return os.Getwd()
	}

	w, err := m.Get(id)
	if err != nil {
		return "", err
	}
	return w.Path, nil
}
