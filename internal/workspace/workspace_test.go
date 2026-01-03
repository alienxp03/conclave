package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager(t *testing.T) {
	// Create temp dir for test config
	tmpDir, err := os.MkdirTemp("", "conclave-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock home dir for testing
	// Since NewManager uses os.UserHomeDir, we can't easily mock it without DI or changing NewManager.
	// But we can create a Manager manually with a custom path.

	manager := &Manager{
		workspaces: make(map[string]*Workspace),
		filePath:   filepath.Join(tmpDir, "workspaces.json"),
	}

	// Test Add
	cwd, _ := os.Getwd()
	w, err := manager.Add("Test Project", cwd)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if w.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", w.Name)
	}
	if w.Path != cwd {
		t.Errorf("Expected path '%s', got '%s'", cwd, w.Path)
	}

	// Test Get
	got, err := manager.Get(w.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != w.ID {
		t.Errorf("Expected ID '%s', got '%s'", w.ID, got.ID)
	}

	// Test List
	list := manager.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(list))
	}

	// Test Delete
	if err := manager.Delete(w.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if len(manager.List()) != 0 {
		t.Errorf("Expected 0 workspaces, got %d", len(manager.List()))
	}
}
