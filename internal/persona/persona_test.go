package persona

import "testing"

func TestDefaultPersonas(t *testing.T) {
	personas := DefaultPersonas()

	if len(personas) != 6 {
		t.Errorf("wrong count: got %d, want 6", len(personas))
	}

	// Check required personas exist
	required := []string{"optimist", "skeptic", "pragmatist", "visionary", "analyst", "devils_advocate"}
	for _, id := range required {
		found := false
		for _, p := range personas {
			if p.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("persona %s not found", id)
		}
	}
}

func TestGet(t *testing.T) {
	t.Run("ExistingPersona", func(t *testing.T) {
		p := Get("optimist")
		if p == nil {
			t.Fatal("persona not found")
		}
		if p.ID != "optimist" {
			t.Errorf("wrong ID: got %s, want optimist", p.ID)
		}
		if p.Name != "Optimist" {
			t.Errorf("wrong Name: got %s, want Optimist", p.Name)
		}
	})

	t.Run("NonexistentPersona", func(t *testing.T) {
		p := Get("nonexistent")
		if p != nil {
			t.Error("expected nil for nonexistent persona")
		}
	})
}

func TestList(t *testing.T) {
	ids := List()
	if len(ids) != 6 {
		t.Errorf("wrong count: got %d, want 6", len(ids))
	}
}

func TestValid(t *testing.T) {
	if !Valid("optimist") {
		t.Error("optimist should be valid")
	}
	if Valid("nonexistent") {
		t.Error("nonexistent should not be valid")
	}
}
