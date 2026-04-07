package service

import (
	"testing"
)

// TestParseCapacityFlowCSV_RealISETFile verifies the real ISET CSV can be parsed
func TestParseCapacityFlowCSV_RealISETFile(t *testing.T) {
	filePath := "/Users/mac/studio/playwright-demo/blueprint/ISET capacity - rack load flow.csv"
	rows, err := ParseCapacityFlowCSV(filePath)
	if err != nil {
		t.Fatalf("ParseCapacityFlowCSV: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected non-empty rows")
	}
	t.Logf("Parsed %d rows from ISET CSV", len(rows))
	
	// Verify first few rows have data
	for i := 0; i < 3 && i < len(rows); i++ {
		t.Logf("Row %d: NodeID=%s Type=%s Variables=%d", i, rows[i].NodeID, rows[i].NodeType, len(rows[i].Variables))
		if rows[i].NodeID == "" {
			t.Errorf("Row %d: empty node_id", i)
		}
		if rows[i].NodeType == "" {
			t.Errorf("Row %d: empty node_type", i)
		}
	}
}

// TestParseCapacityFlowCSV_RealISETFile_VariableCount verifies variable parsing is complete
func TestParseCapacityFlowCSV_RealISETFile_VariableCount(t *testing.T) {
	filePath := "/Users/mac/studio/playwright-demo/blueprint/ISET capacity - rack load flow.csv"
	rows, err := ParseCapacityFlowCSV(filePath)
	if err != nil {
		t.Fatalf("ParseCapacityFlowCSV: %v", err)
	}

	// Count by node type
	typeCount := make(map[string]int)
	totalVars := 0
	for _, row := range rows {
		typeCount[row.NodeType]++
		totalVars += len(row.Variables)
	}

	t.Logf("Summary: %d total rows, %d total variables", len(rows), totalVars)
	for nodeType, count := range typeCount {
		t.Logf("  %s: %d nodes", nodeType, count)
	}

	if totalVars == 0 {
		t.Fatal("expected non-zero total variables")
	}
	if len(typeCount) == 0 {
		t.Fatal("expected at least one node type")
	}
}

// TestParseCapacityFlowCSV_EdgeCases verifies error handling
func TestParseCapacityFlowCSV_EdgeCases(t *testing.T) {
	// Non-existent file
	_, err := ParseCapacityFlowCSV("/nonexistent/path.csv")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	t.Logf("Non-existent file error (expected): %v", err)
}
