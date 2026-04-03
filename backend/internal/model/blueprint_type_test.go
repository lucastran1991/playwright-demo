package model

import (
	"testing"
	"time"
)

// TestBlueprintTypeStructure tests the BlueprintType model structure.
func TestBlueprintTypeStructure(t *testing.T) {
	now := time.Now()
	bt := BlueprintType{
		ID:         1,
		Name:       "Cooling System",
		Slug:       "cooling-system",
		FolderName: "Cooling system_Blueprint",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if bt.ID != 1 {
		t.Errorf("ID mismatch: expected 1, got %d", bt.ID)
	}

	if bt.Name != "Cooling System" {
		t.Errorf("Name mismatch: expected 'Cooling System', got '%s'", bt.Name)
	}

	if bt.Slug != "cooling-system" {
		t.Errorf("Slug mismatch: expected 'cooling-system', got '%s'", bt.Slug)
	}

	if bt.FolderName != "Cooling system_Blueprint" {
		t.Errorf("FolderName mismatch: expected 'Cooling system_Blueprint', got '%s'", bt.FolderName)
	}

	if !bt.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt mismatch")
	}

	if !bt.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt mismatch")
	}
}

// TestBlueprintTypeZeroValue tests zero value initialization.
func TestBlueprintTypeZeroValue(t *testing.T) {
	bt := BlueprintType{}

	if bt.ID != 0 {
		t.Errorf("Expected zero ID, got %d", bt.ID)
	}

	if bt.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", bt.Name)
	}

	if bt.Slug != "" {
		t.Errorf("Expected empty Slug, got '%s'", bt.Slug)
	}
}
