package model

import (
	"testing"
	"time"
)

// TestBlueprintNodeStructure tests the BlueprintNode model structure.
func TestBlueprintNodeStructure(t *testing.T) {
	now := time.Now()
	bn := BlueprintNode{
		ID:        42,
		NodeID:    "RACK-R1-Z1-01",
		Name:      "Rack 1 in Zone 1",
		NodeType:  "Rack",
		NodeRole:  "physical",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if bn.ID != 42 {
		t.Errorf("ID mismatch: expected 42, got %d", bn.ID)
	}

	if bn.NodeID != "RACK-R1-Z1-01" {
		t.Errorf("NodeID mismatch: expected 'RACK-R1-Z1-01', got '%s'", bn.NodeID)
	}

	if bn.Name != "Rack 1 in Zone 1" {
		t.Errorf("Name mismatch: expected 'Rack 1 in Zone 1', got '%s'", bn.Name)
	}

	if bn.NodeType != "Rack" {
		t.Errorf("NodeType mismatch: expected 'Rack', got '%s'", bn.NodeType)
	}

	if bn.NodeRole != "physical" {
		t.Errorf("NodeRole mismatch: expected 'physical', got '%s'", bn.NodeRole)
	}
}

// TestBlueprintNodeZeroValue tests zero value initialization.
func TestBlueprintNodeZeroValue(t *testing.T) {
	bn := BlueprintNode{}

	if bn.ID != 0 {
		t.Errorf("Expected zero ID, got %d", bn.ID)
	}

	if bn.NodeID != "" {
		t.Errorf("Expected empty NodeID, got '%s'", bn.NodeID)
	}

	if bn.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", bn.Name)
	}

	if bn.NodeRole != "" {
		t.Errorf("Expected empty NodeRole, got '%s'", bn.NodeRole)
	}
}

// TestBlueprintNodeOptionalFields tests that NodeRole is optional.
func TestBlueprintNodeOptionalFields(t *testing.T) {
	bn := BlueprintNode{
		ID:       1,
		NodeID:   "TEST-01",
		Name:     "Test Node",
		NodeType: "Test",
	}

	// NodeRole can be omitted
	if bn.NodeRole != "" {
		t.Errorf("Expected empty NodeRole for optional field, got '%s'", bn.NodeRole)
	}
}
