package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// NodeRow represents a parsed row from a blueprint Nodes CSV file.
type NodeRow struct {
	NodeID   string
	Name     string
	Role     string
	OrgPath  string
	NodeType string
}

// EdgeRow represents a parsed row from a blueprint Edges CSV file.
type EdgeRow struct {
	FromName    string
	FromNodeID  string
	FromOrgPath string
	ToName      string
	ToNodeID    string
	ToOrgPath   string
}

// DomainFolder represents a discovered blueprint domain directory.
type DomainFolder struct {
	Name string // folder name, e.g. "Cooling system_Blueprint"
	Path string // full path to the folder
}

// DiscoverDomains scans basePath for subdirectories representing blueprint domains.
func DiscoverDomains(basePath string) ([]DomainFolder, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprint directory %s: %w", basePath, err)
	}

	var domains []DomainFolder
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		domains = append(domains, DomainFolder{
			Name: e.Name(),
			Path: filepath.Join(basePath, e.Name()),
		})
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })
	return domains, nil
}

// FindCSVFile finds the nodes or edges CSV file in a domain folder.
// It looks for files containing "Node" or "Edge" (case-insensitive) with .csv extension.
func FindCSVFile(domainPath, kind string) (string, error) {
	entries, err := os.ReadDir(domainPath)
	if err != nil {
		return "", fmt.Errorf("failed to read domain directory %s: %w", domainPath, err)
	}
	for _, e := range entries {
		lower := strings.ToLower(e.Name())
		if strings.HasSuffix(lower, ".csv") && strings.Contains(lower, strings.ToLower(kind)) {
			return filepath.Join(domainPath, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no %s CSV file found in %s", kind, domainPath)
}

// ParseNodesCSV reads and parses a blueprint Nodes CSV file.
func ParseNodesCSV(filePath string) ([]NodeRow, error) {
	records, err := ReadCSV(filePath)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file: %s", filePath)
	}

	// Validate header
	header := records[0]
	if len(header) < 5 {
		return nil, fmt.Errorf("invalid nodes CSV header in %s: expected 5 columns, got %d", filePath, len(header))
	}

	var rows []NodeRow
	for _, rec := range records[1:] {
		if len(rec) < 5 || strings.TrimSpace(rec[0]) == "" {
			continue
		}
		rows = append(rows, NodeRow{
			NodeID:   strings.TrimSpace(rec[0]),
			Name:     strings.TrimSpace(rec[1]),
			Role:     strings.TrimSpace(rec[2]),
			OrgPath:  strings.TrimSpace(rec[3]),
			NodeType: strings.TrimSpace(rec[4]),
		})
	}
	return rows, nil
}

// ParseEdgesCSV reads and parses a blueprint Edges CSV file.
func ParseEdgesCSV(filePath string) ([]EdgeRow, error) {
	records, err := ReadCSV(filePath)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file: %s", filePath)
	}

	header := records[0]
	if len(header) < 6 {
		return nil, fmt.Errorf("invalid edges CSV header in %s: expected 6 columns, got %d", filePath, len(header))
	}

	var rows []EdgeRow
	for _, rec := range records[1:] {
		if len(rec) < 6 || strings.TrimSpace(rec[1]) == "" || strings.TrimSpace(rec[4]) == "" {
			continue
		}
		rows = append(rows, EdgeRow{
			FromName:    strings.TrimSpace(rec[0]),
			FromNodeID:  strings.TrimSpace(rec[1]),
			FromOrgPath: strings.TrimSpace(rec[2]),
			ToName:      strings.TrimSpace(rec[3]),
			ToNodeID:    strings.TrimSpace(rec[4]),
			ToOrgPath:   strings.TrimSpace(rec[5]),
		})
	}
	return rows, nil
}

// FolderToSlug converts a blueprint folder name to a URL-friendly slug.
// Example: "Cooling system_Blueprint" -> "cooling-system"
func FolderToSlug(folder string) string {
	s := strings.TrimSuffix(folder, "_Blueprint")
	s = strings.TrimSuffix(s, " Blueprint")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// FolderToName converts a blueprint folder name to a display name.
// Example: "Cooling system_Blueprint" -> "Cooling System"
func FolderToName(folder string) string {
	s := strings.TrimSuffix(folder, "_Blueprint")
	s = strings.TrimSuffix(s, " Blueprint")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}

func ReadCSV(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %s: %w", filePath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // allow variable field count
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file %s: %w", filePath, err)
	}
	return records, nil
}
