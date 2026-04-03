package service

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseCapacityNodesCSV_Valid tests parsing of a valid Capacity Nodes CSV file.
func TestParseCapacityNodesCSV_Valid(t *testing.T) {
	csvFile := filepath.Join("../../testdata/models", "capacity-nodes-valid.csv")

	rows, err := ParseCapacityNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseCapacityNodesCSV failed: %v", err)
	}

	if len(rows) != 6 {
		t.Errorf("Expected 6 rows, got %d", len(rows))
	}

	// Test first row - all false
	if rows[0].NodeType != "Rack PDU" {
		t.Errorf("Expected NodeType 'Rack PDU', got '%s'", rows[0].NodeType)
	}
	if rows[0].Topology != "Electrical System" {
		t.Errorf("Expected Topology 'Electrical System', got '%s'", rows[0].Topology)
	}
	if rows[0].IsCapacityNode != false {
		t.Errorf("Expected IsCapacityNode false, got %v", rows[0].IsCapacityNode)
	}
	if rows[0].ActiveConstraint != false {
		t.Errorf("Expected ActiveConstraint false, got %v", rows[0].ActiveConstraint)
	}

	// Test second row - both true
	if rows[1].NodeType != "RPP" {
		t.Errorf("Expected NodeType 'RPP', got '%s'", rows[1].NodeType)
	}
	if rows[1].IsCapacityNode != true {
		t.Errorf("Expected IsCapacityNode true, got %v", rows[1].IsCapacityNode)
	}
	if rows[1].ActiveConstraint != true {
		t.Errorf("Expected ActiveConstraint true, got %v", rows[1].ActiveConstraint)
	}

	// Test row with mixed booleans
	if rows[4].NodeType != "Rack" {
		t.Errorf("Expected NodeType 'Rack', got '%s'", rows[4].NodeType)
	}
	if rows[4].IsCapacityNode != true {
		t.Errorf("Expected IsCapacityNode true for Rack, got %v", rows[4].IsCapacityNode)
	}
	if rows[4].ActiveConstraint != false {
		t.Errorf("Expected ActiveConstraint false for Rack, got %v", rows[4].ActiveConstraint)
	}
}

// TestParseCapacityNodesCSV_EmptyFile tests handling of empty CSV files.
func TestParseCapacityNodesCSV_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "empty.csv")

	if err := os.WriteFile(csvFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseCapacityNodesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for empty CSV file")
	}
}

// TestParseCapacityNodesCSV_BadHeader tests handling of invalid headers.
func TestParseCapacityNodesCSV_BadHeader(t *testing.T) {
	csvFile := filepath.Join("../../testdata/models", "capacity-nodes-bad-header.csv")

	_, err := ParseCapacityNodesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for invalid header")
	}
}

// TestParseCapacityNodesCSV_OnlyHeaders tests CSV with only headers, no data rows.
func TestParseCapacityNodesCSV_OnlyHeaders(t *testing.T) {
	csvFile := filepath.Join("../../testdata/models", "capacity-nodes-empty.csv")

	rows, err := ParseCapacityNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseCapacityNodesCSV failed: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

// TestParseCapacityNodesCSV_MissingNodeType tests skipping rows with empty node type.
func TestParseCapacityNodesCSV_MissingNodeType(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "missing_node_type.csv")

	content := `Node Type,Topology,Capacity Node (Capacity Domain),ActiveConstraint
,Electrical System,True,True
RPP,Electrical System,True,True
  ,Cooling System,False,False
Rack,Spatial Topology,True,False
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseCapacityNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseCapacityNodesCSV failed: %v", err)
	}

	// Should skip rows with empty or whitespace-only NodeType
	if len(rows) != 2 {
		t.Errorf("Expected 2 valid rows (skipped 2 with empty NodeType), got %d", len(rows))
	}

	if rows[0].NodeType != "RPP" {
		t.Errorf("Expected first NodeType 'RPP', got '%s'", rows[0].NodeType)
	}
	if rows[1].NodeType != "Rack" {
		t.Errorf("Expected second NodeType 'Rack', got '%s'", rows[1].NodeType)
	}
}

// TestParseCapacityNodesCSV_WhitespaceHandling tests that whitespace is properly trimmed.
func TestParseCapacityNodesCSV_WhitespaceHandling(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "whitespace.csv")

	content := `Node Type,Topology,Capacity Node (Capacity Domain),ActiveConstraint
  RPP  ,  Electrical System  , True , False
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseCapacityNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseCapacityNodesCSV failed: %v", err)
	}

	if rows[0].NodeType != "RPP" {
		t.Errorf("Whitespace not trimmed from NodeType: '%s'", rows[0].NodeType)
	}
	if rows[0].Topology != "Electrical System" {
		t.Errorf("Whitespace not trimmed from Topology: '%s'", rows[0].Topology)
	}
}

// TestParseCapacityNodesCSV_NonExistentFile tests error for missing file.
func TestParseCapacityNodesCSV_NonExistentFile(t *testing.T) {
	_, err := ParseCapacityNodesCSV("/nonexistent/capacity-nodes.csv")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

// TestParseBoolStr tests the parseBoolStr helper function.
func TestParseBoolStr(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"True", true},
		{"true", true},
		{"TRUE", true},
		{"tRuE", true},
		{"False", false},
		{"false", false},
		{"FALSE", false},
		{"fAlSe", false},
		{"", false},
		{"yes", false},
		{"no", false},
		{"1", false},
		{"0", false},
		{"  true  ", true},
		{"  false  ", false},
	}

	for _, test := range tests {
		result := parseBoolStr(test.input)
		if result != test.expected {
			t.Errorf("parseBoolStr(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

// TestParseDependenciesCSV_Valid tests parsing of a valid Dependencies CSV file.
func TestParseDependenciesCSV_Valid(t *testing.T) {
	csvFile := filepath.Join("../../testdata/models", "dependencies-valid.csv")

	rows, err := ParseDependenciesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseDependenciesCSV failed: %v", err)
	}

	if len(rows) != 5 {
		t.Errorf("Expected 5 rows, got %d", len(rows))
	}

	// Test first row with upstream level
	if rows[0].NodeType != "Rack" {
		t.Errorf("Expected NodeType 'Rack', got '%s'", rows[0].NodeType)
	}
	if rows[0].DependencyNodeType != "RPP" {
		t.Errorf("Expected DependencyNodeType 'RPP', got '%s'", rows[0].DependencyNodeType)
	}
	if rows[0].TopologicalRelationship != "Upstream" {
		t.Errorf("Expected TopologicalRelationship 'Upstream', got '%s'", rows[0].TopologicalRelationship)
	}
	if rows[0].UpstreamLevel == nil || *rows[0].UpstreamLevel != 1 {
		t.Errorf("Expected UpstreamLevel 1, got %v", rows[0].UpstreamLevel)
	}

	// Test row with local relationship and empty level
	if rows[3].NodeType != "Rack" {
		t.Errorf("Expected NodeType 'Rack', got '%s'", rows[3].NodeType)
	}
	if rows[3].DependencyNodeType != "RDHx" {
		t.Errorf("Expected DependencyNodeType 'RDHx', got '%s'", rows[3].DependencyNodeType)
	}
	if rows[3].TopologicalRelationship != "Local" {
		t.Errorf("Expected TopologicalRelationship 'Local', got '%s'", rows[3].TopologicalRelationship)
	}
	if rows[3].UpstreamLevel != nil {
		t.Errorf("Expected UpstreamLevel nil for Local relationship, got %v", rows[3].UpstreamLevel)
	}

	// Test Row with Upstream relationship
	if rows[4].NodeType != "Row" {
		t.Errorf("Expected NodeType 'Row', got '%s'", rows[4].NodeType)
	}
	if rows[4].TopologicalRelationship != "Upstream" {
		t.Errorf("Expected TopologicalRelationship 'Upstream', got '%s'", rows[4].TopologicalRelationship)
	}
	if rows[4].UpstreamLevel == nil || *rows[4].UpstreamLevel != 1 {
		t.Errorf("Expected UpstreamLevel 1, got %v", rows[4].UpstreamLevel)
	}
}

// TestParseDependenciesCSV_EmptyUpstreamLevel tests parsing with empty upstream level field.
func TestParseDependenciesCSV_EmptyUpstreamLevel(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "deps_empty_level.csv")

	content := `Node Type,Dependency Node Type,Relationship Type,Topological Relationship,Upstream Level
Rack,RDHx,Dependency,Local,
Rack,RPP,Dependency,Upstream,1
Row,Air Zone,Dependency,Upstream,
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseDependenciesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseDependenciesCSV failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	// First row: empty level
	if rows[0].UpstreamLevel != nil {
		t.Errorf("Expected UpstreamLevel nil for empty field, got %v", rows[0].UpstreamLevel)
	}

	// Second row: level = 1
	if rows[1].UpstreamLevel == nil || *rows[1].UpstreamLevel != 1 {
		t.Errorf("Expected UpstreamLevel 1, got %v", rows[1].UpstreamLevel)
	}

	// Third row: empty level
	if rows[2].UpstreamLevel != nil {
		t.Errorf("Expected UpstreamLevel nil for empty field, got %v", rows[2].UpstreamLevel)
	}
}

// TestParseDependenciesCSV_BadHeader tests handling of invalid headers.
func TestParseDependenciesCSV_BadHeader(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "deps_bad_header.csv")

	content := "Col1,Col2,Col3,Col4\nval1,val2,val3,val4"

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseDependenciesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for invalid header")
	}
}

// TestParseDependenciesCSV_MissingRequiredFields tests handling of missing required fields.
func TestParseDependenciesCSV_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "deps_missing_fields.csv")

	content := `Node Type,Dependency Node Type,Relationship Type,Topological Relationship,Upstream Level
,RPP,Dependency,Upstream,1
Rack,,Dependency,Upstream,1
Rack,Room PDU,Dependency,Upstream,2
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseDependenciesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseDependenciesCSV failed: %v", err)
	}

	// Should skip rows with empty NodeType or DependencyNodeType
	if len(rows) != 1 {
		t.Errorf("Expected 1 valid row (skipped 2 with missing fields), got %d", len(rows))
	}

	if rows[0].NodeType != "Rack" || rows[0].DependencyNodeType != "Room PDU" {
		t.Errorf("Expected Rack->Room PDU, got %s->%s", rows[0].NodeType, rows[0].DependencyNodeType)
	}
}

// TestParseOptionalInt tests the parseOptionalInt helper function.
func TestParseOptionalInt(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"1", toPtr(1)},
		{"42", toPtr(42)},
		{"0", toPtr(0)},
		{"-5", toPtr(-5)},
		{"", nil},
		{"  ", nil},
		{"abc", nil},
		{"12.5", nil},
		{"  123  ", toPtr(123)},
	}

	for _, test := range tests {
		result := parseOptionalInt(test.input)
		if (result == nil && test.expected != nil) || (result != nil && test.expected == nil) {
			t.Errorf("parseOptionalInt(%q) = %v, expected %v", test.input, result, test.expected)
		} else if result != nil && test.expected != nil && *result != *test.expected {
			t.Errorf("parseOptionalInt(%q) = %d, expected %d", test.input, *result, *test.expected)
		}
	}
}

// TestParseImpactsCSV_Valid tests parsing of a valid Impacts CSV file.
func TestParseImpactsCSV_Valid(t *testing.T) {
	csvFile := filepath.Join("../../testdata/models", "impacts-valid.csv")

	rows, err := ParseImpactsCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseImpactsCSV failed: %v", err)
	}

	if len(rows) != 5 {
		t.Errorf("Expected 5 rows, got %d", len(rows))
	}

	// Test first row - Load relationship with empty level
	if rows[0].NodeType != "Rack PDU" {
		t.Errorf("Expected NodeType 'Rack PDU', got '%s'", rows[0].NodeType)
	}
	if rows[0].ImpactNodeType != "Rack" {
		t.Errorf("Expected ImpactNodeType 'Rack', got '%s'", rows[0].ImpactNodeType)
	}
	if rows[0].TopologicalRelationship != "Load" {
		t.Errorf("Expected TopologicalRelationship 'Load', got '%s'", rows[0].TopologicalRelationship)
	}
	if rows[0].DownstreamLevel != nil {
		t.Errorf("Expected DownstreamLevel nil, got %v", rows[0].DownstreamLevel)
	}

	// Test row with downstream level
	if rows[1].NodeType != "RPP" {
		t.Errorf("Expected NodeType 'RPP', got '%s'", rows[1].NodeType)
	}
	if rows[1].TopologicalRelationship != "Downstream" {
		t.Errorf("Expected TopologicalRelationship 'Downstream', got '%s'", rows[1].TopologicalRelationship)
	}
	if rows[1].DownstreamLevel == nil || *rows[1].DownstreamLevel != 1 {
		t.Errorf("Expected DownstreamLevel 1, got %v", rows[1].DownstreamLevel)
	}
}

// TestParseImpactsCSV_EmptyDownstreamLevel tests parsing with empty downstream level field.
func TestParseImpactsCSV_EmptyDownstreamLevel(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "impacts_empty_level.csv")

	content := `Node Type,Impact Node Type,Topological Relationship,Downstream Level
Rack PDU,Rack,Load,
RPP,Rack PDU,Downstream,1
UPS,Rack,Load,
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseImpactsCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseImpactsCSV failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	// First row: empty level
	if rows[0].DownstreamLevel != nil {
		t.Errorf("Expected DownstreamLevel nil for empty field, got %v", rows[0].DownstreamLevel)
	}

	// Second row: level = 1
	if rows[1].DownstreamLevel == nil || *rows[1].DownstreamLevel != 1 {
		t.Errorf("Expected DownstreamLevel 1, got %v", rows[1].DownstreamLevel)
	}
}

// TestParseImpactsCSV_BadHeader tests handling of invalid headers.
func TestParseImpactsCSV_BadHeader(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "impacts_bad_header.csv")

	content := "Col1,Col2,Col3\nval1,val2,val3"

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseImpactsCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for invalid header")
	}
}

// TestParseImpactsCSV_MissingRequiredFields tests handling of missing required fields.
func TestParseImpactsCSV_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "impacts_missing_fields.csv")

	content := `Node Type,Impact Node Type,Topological Relationship,Downstream Level
,Rack,Load,
RPP,,Downstream,1
RPP,Rack PDU,Downstream,2
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseImpactsCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseImpactsCSV failed: %v", err)
	}

	// Should skip rows with empty NodeType or ImpactNodeType
	if len(rows) != 1 {
		t.Errorf("Expected 1 valid row (skipped 2 with missing fields), got %d", len(rows))
	}

	if rows[0].NodeType != "RPP" || rows[0].ImpactNodeType != "Rack PDU" {
		t.Errorf("Expected RPP->Rack PDU, got %s->%s", rows[0].NodeType, rows[0].ImpactNodeType)
	}
}

// TestParseImpactsCSV_NonExistentFile tests error for missing file.
func TestParseImpactsCSV_NonExistentFile(t *testing.T) {
	_, err := ParseImpactsCSV("/nonexistent/impacts.csv")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

// Helper function to create int pointers for test comparisons
func toPtr(i int) *int {
	return &i
}
