package service

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDiscoverDomains tests domain discovery from a directory structure.
func TestDiscoverDomains(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test domain directories
	domainDirs := []string{
		filepath.Join(tmpDir, "Domain_A_Blueprint"),
		filepath.Join(tmpDir, "Domain_B_Blueprint"),
		filepath.Join(tmpDir, ".hidden"),
		filepath.Join(tmpDir, "not_a_domain.txt"),
	}

	for _, d := range domainDirs[:2] {
		if err := os.Mkdir(d, 0755); err != nil {
			t.Fatalf("Failed to create test dir: %v", err)
		}
	}

	// Create a file (should be skipped)
	if f, err := os.Create(filepath.Join(tmpDir, "not_a_domain.txt")); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	} else {
		f.Close()
	}

	// Create hidden directory (should be skipped)
	if err := os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0755); err != nil {
		t.Fatalf("Failed to create hidden dir: %v", err)
	}

	domains, err := DiscoverDomains(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverDomains failed: %v", err)
	}

	if len(domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains))
	}

	expectedNames := map[string]bool{
		"Domain_A_Blueprint": true,
		"Domain_B_Blueprint": true,
	}

	for _, d := range domains {
		if !expectedNames[d.Name] {
			t.Errorf("Unexpected domain: %s", d.Name)
		}
		if d.Path != filepath.Join(tmpDir, d.Name) {
			t.Errorf("Domain path mismatch: expected %s, got %s", filepath.Join(tmpDir, d.Name), d.Path)
		}
	}
}

// TestDiscoverDomains_NonExistentPath tests error handling for non-existent paths.
func TestDiscoverDomains_NonExistentPath(t *testing.T) {
	domains, err := DiscoverDomains("/nonexistent/path")
	if err == nil {
		t.Fatal("Expected error for non-existent path")
	}
	if domains != nil {
		t.Errorf("Expected nil domains, got %v", domains)
	}
}

// TestFolderToSlug tests slug generation from folder names.
func TestFolderToSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cooling system_Blueprint", "cooling-system"},
		{"Cooling system Blueprint", "cooling-system"},
		{"Electrical_System_Blueprint", "electrical-system"},
		{"Simple Blueprint", "simple"},
		{"Simple_Blueprint", "simple"},
		{"Multi Word System_Blueprint", "multi-word-system"},
		{"Already-slug", "already-slug"},
		{" Spaces Around _Blueprint", "spaces-around"},
		{"Multiple___Underscores_Blueprint", "multiple-underscores"},
		{"", ""},
	}

	for _, test := range tests {
		result := FolderToSlug(test.input)
		if result != test.expected {
			t.Errorf("FolderToSlug(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestFolderToName tests name generation from folder names.
func TestFolderToName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cooling system_Blueprint", "Cooling system"},
		{"Cooling system Blueprint", "Cooling system"},
		{"Electrical_System_Blueprint", "Electrical System"},
		{"Simple Blueprint", "Simple"},
		{"Simple_Blueprint", "Simple"},
		{"Multi Word System_Blueprint", "Multi Word System"},
	}

	for _, test := range tests {
		result := FolderToName(test.input)
		if result != test.expected {
			t.Errorf("FolderToName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestParseNodesCSV tests parsing of a valid Nodes CSV file.
func TestParseNodesCSV_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test_nodes.csv")

	// Create test CSV
	content := `Node ID,Node Name,Node Role,Org Path,Node Type
ROOT-01,Root Node,admin,/,Root
CHILD-01,Child Node 1,,/Root,Child
CHILD-02,Child Node 2,,/Root,Child
LEAF-01,Leaf Node,operator,/Root/Child1,Leaf
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	rows, err := ParseNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseNodesCSV failed: %v", err)
	}

	if len(rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(rows))
	}

	// Test first row
	if rows[0].NodeID != "ROOT-01" {
		t.Errorf("Expected NodeID 'ROOT-01', got '%s'", rows[0].NodeID)
	}
	if rows[0].Name != "Root Node" {
		t.Errorf("Expected Name 'Root Node', got '%s'", rows[0].Name)
	}
	if rows[0].Role != "admin" {
		t.Errorf("Expected Role 'admin', got '%s'", rows[0].Role)
	}

	// Test row with empty role
	if rows[1].Role != "" {
		t.Errorf("Expected empty Role for CHILD-01, got '%s'", rows[1].Role)
	}
}

// TestParseNodesCSV_EmptyFile tests handling of empty CSV files.
func TestParseNodesCSV_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "empty.csv")

	if err := os.WriteFile(csvFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseNodesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for empty CSV file")
	}
}

// TestParseNodesCSV_BadHeader tests handling of invalid headers.
func TestParseNodesCSV_BadHeader(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "bad_header.csv")

	// Only 3 columns instead of 5
	content := "Col1,Col2,Col3\nval1,val2,val3"

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseNodesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for invalid header")
	}
}

// TestParseNodesCSV_MalformedRows tests handling of malformed rows.
func TestParseNodesCSV_MalformedRows(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "malformed.csv")

	content := `Node ID,Node Name,Node Role,Org Path,Node Type
ROOT-01,Root Node,admin,/,Root
,BadRow,missing,id,Field
CHILD-01,Valid,role,path,Type
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseNodesCSV failed: %v", err)
	}

	// Should skip empty node ID row and only get 2 valid rows
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows (skipped malformed), got %d", len(rows))
	}
}

// TestParseNodesCSV_WhitespaceHandling tests whitespace trimming.
func TestParseNodesCSV_WhitespaceHandling(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "whitespace.csv")

	content := `Node ID,Node Name,Node Role,Org Path,Node Type
  ROOT-01  ,  Root Node  , admin , / , Root
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseNodesCSV failed: %v", err)
	}

	if rows[0].NodeID != "ROOT-01" {
		t.Errorf("Whitespace not trimmed from NodeID: '%s'", rows[0].NodeID)
	}
	if rows[0].Name != "Root Node" {
		t.Errorf("Whitespace not trimmed from Name: '%s'", rows[0].Name)
	}
}

// TestParseEdgesCSV_Valid tests parsing of a valid Edges CSV file.
func TestParseEdgesCSV_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test_edges.csv")

	content := `From Node Name,From Node ID,From Node Org Path,To Node Name,To Node ID,To Node Org Path
Root Node,ROOT-01,/,Child Node 1,CHILD-01,/Root
Root Node,ROOT-01,/,Child Node 2,CHILD-02,/Root
Child Node 1,CHILD-01,/Root,Leaf Node,LEAF-01,/Root/Child1
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	rows, err := ParseEdgesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseEdgesCSV failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	// Test first edge
	if rows[0].FromNodeID != "ROOT-01" {
		t.Errorf("Expected FromNodeID 'ROOT-01', got '%s'", rows[0].FromNodeID)
	}
	if rows[0].ToNodeID != "CHILD-01" {
		t.Errorf("Expected ToNodeID 'CHILD-01', got '%s'", rows[0].ToNodeID)
	}
}

// TestParseEdgesCSV_EmptyFile tests handling of empty Edges CSV.
func TestParseEdgesCSV_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "empty_edges.csv")

	if err := os.WriteFile(csvFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseEdgesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for empty CSV file")
	}
}

// TestParseEdgesCSV_BadHeader tests handling of invalid edge headers.
func TestParseEdgesCSV_BadHeader(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "bad_edge_header.csv")

	// Only 4 columns instead of 6
	content := "Col1,Col2,Col3,Col4\nval1,val2,val3,val4"

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseEdgesCSV(csvFile)
	if err == nil {
		t.Fatal("Expected error for invalid edge header")
	}
}

// TestParseEdgesCSV_MissingRequiredFields tests handling of missing required fields.
func TestParseEdgesCSV_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "missing_fields.csv")

	// Row 1: Missing FromNodeID (rec[1] empty) - should skip
	// Row 2: Missing ToNodeID (rec[4] empty) - should skip
	// Row 3: Both IDs present - should include
	// Row 4: Both IDs present - should include
	content := `From Node Name,From Node ID,From Node Org Path,To Node Name,To Node ID,To Node Org Path
Node1,,/,Node2,TO-01,/
Node3,FROM-01,/,Node4,,/
Node5,FROM-02,/,Node6,TO-03,/
Node7,FROM-03,/,Node8,TO-04,/
`

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseEdgesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseEdgesCSV failed: %v", err)
	}

	// The validation checks: rec[1] (FromNodeID) and rec[4] (ToNodeID) are both non-empty
	// Row 1: rec[1]="" -> skip
	// Row 2: rec[4]="" -> skip
	// Row 3: rec[1]="FROM-02", rec[4]="TO-03" -> include
	// Row 4: rec[1]="FROM-03", rec[4]="TO-04" -> include
	if len(rows) != 2 {
		t.Errorf("Expected 2 valid rows (skipped 2 with missing fields), got %d", len(rows))
	}

	if len(rows) > 0 && rows[0].FromNodeID != "FROM-02" {
		t.Errorf("Expected first FromNodeID 'FROM-02', got '%s'", rows[0].FromNodeID)
	}
}

// TestFindCSVFile_NodesFound tests finding a Nodes CSV file.
func TestFindCSVFile_NodesFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various files
	nodesFile := filepath.Join(tmpDir, "Blueprint_Nodes.csv")
	if err := os.WriteFile(nodesFile, []byte("header"), 0644); err != nil {
		t.Fatalf("Failed to create nodes file: %v", err)
	}

	// Create unrelated file
	otherFile := filepath.Join(tmpDir, "other.txt")
	if err := os.WriteFile(otherFile, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	found, err := FindCSVFile(tmpDir, "Node")
	if err != nil {
		t.Fatalf("FindCSVFile failed: %v", err)
	}

	if found != nodesFile {
		t.Errorf("Expected %s, got %s", nodesFile, found)
	}
}

// TestFindCSVFile_NotFound tests error when CSV file is not found.
func TestFindCSVFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindCSVFile(tmpDir, "Node")
	if err == nil {
		t.Fatal("Expected error when CSV file not found")
	}
}

// TestFindCSVFile_CaseInsensitive tests case-insensitive file matching.
func TestFindCSVFile_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with different case
	edgesFile := filepath.Join(tmpDir, "data_EDGES.csv")
	if err := os.WriteFile(edgesFile, []byte("header"), 0644); err != nil {
		t.Fatalf("Failed to create edges file: %v", err)
	}

	found, err := FindCSVFile(tmpDir, "Edge")
	if err != nil {
		t.Fatalf("FindCSVFile failed: %v", err)
	}

	if found != edgesFile {
		t.Errorf("Expected %s, got %s", edgesFile, found)
	}
}

// TestParseNodesCSV_NonExistentFile tests error for missing file.
func TestParseNodesCSV_NonExistentFile(t *testing.T) {
	_, err := ParseNodesCSV("/nonexistent/file.csv")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

// TestParseEdgesCSV_NonExistentFile tests error for missing edges file.
func TestParseEdgesCSV_NonExistentFile(t *testing.T) {
	_, err := ParseEdgesCSV("/nonexistent/edges.csv")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

// TestParseNodesCSV_LargeFile tests parsing a larger CSV to ensure performance.
func TestParseNodesCSV_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "large.csv")

	// Create CSV with 1000 rows
	content := "Node ID,Node Name,Node Role,Org Path,Node Type\n"
	for i := 1; i <= 1000; i++ {
		content += "NODE-" + string(rune('0'+i%10)) + ",Node " + string(rune('A'+i%26)) + ",role,path," + string(rune('T'+i%5)) + "\n"
	}

	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rows, err := ParseNodesCSV(csvFile)
	if err != nil {
		t.Fatalf("ParseNodesCSV failed: %v", err)
	}

	if len(rows) != 1000 {
		t.Errorf("Expected 1000 rows, got %d", len(rows))
	}
}
