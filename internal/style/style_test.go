package style

import "testing"

func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	if len(styles) != 4 {
		t.Errorf("wrong count: got %d, want 4", len(styles))
	}

	// Check required styles exist
	required := []string{"adversarial", "collaborative", "analytical", "socratic"}
	for _, id := range required {
		found := false
		for _, s := range styles {
			if s.ID == id {
				found = true
				if s.OpeningPrompt == "" {
					t.Errorf("style %s has empty OpeningPrompt", id)
				}
				if s.ResponsePrompt == "" {
					t.Errorf("style %s has empty ResponsePrompt", id)
				}
				if s.ConclusionPrompt == "" {
					t.Errorf("style %s has empty ConclusionPrompt", id)
				}
				break
			}
		}
		if !found {
			t.Errorf("style %s not found", id)
		}
	}
}

func TestGet(t *testing.T) {
	t.Run("ExistingStyle", func(t *testing.T) {
		s := Get("collaborative")
		if s == nil {
			t.Fatal("style not found")
		}
		if s.ID != "collaborative" {
			t.Errorf("wrong ID: got %s, want collaborative", s.ID)
		}
	})

	t.Run("NonexistentStyle", func(t *testing.T) {
		s := Get("nonexistent")
		if s != nil {
			t.Error("expected nil for nonexistent style")
		}
	})
}

func TestList(t *testing.T) {
	ids := List()
	if len(ids) != 4 {
		t.Errorf("wrong count: got %d, want 4", len(ids))
	}
}

func TestValid(t *testing.T) {
	if !Valid("collaborative") {
		t.Error("collaborative should be valid")
	}
	if Valid("nonexistent") {
		t.Error("nonexistent should not be valid")
	}
}

func TestDefault(t *testing.T) {
	d := Default()
	if d == nil {
		t.Fatal("default style is nil")
	}
	if d.ID != "collaborative" {
		t.Errorf("wrong default style: got %s, want collaborative", d.ID)
	}
}
