package service

import (
	"fmt"
	"strconv"
	"strings"
)

// CapacityNodeTypeRow represents a parsed row from Capacity Nodes.csv.
type CapacityNodeTypeRow struct {
	NodeType         string
	Topology         string
	IsCapacityNode   bool
	ActiveConstraint bool
}

// DependencyRuleRow represents a parsed row from Dependencies.csv.
type DependencyRuleRow struct {
	NodeType                string
	DependencyNodeType      string
	RelationshipType        string
	TopologicalRelationship string
	UpstreamLevel           *int
}

// ImpactRuleRow represents a parsed row from Impacts.csv.
type ImpactRuleRow struct {
	NodeType                string
	ImpactNodeType          string
	TopologicalRelationship string
	DownstreamLevel         *int
}

// ParseCapacityNodesCSV parses the Capacity Nodes CSV file.
func ParseCapacityNodesCSV(filePath string) ([]CapacityNodeTypeRow, error) {
	records, err := ReadCSV(filePath)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file: %s", filePath)
	}
	if len(records[0]) < 4 {
		return nil, fmt.Errorf("invalid capacity nodes CSV header in %s: expected 4 columns, got %d", filePath, len(records[0]))
	}

	var rows []CapacityNodeTypeRow
	for _, rec := range records[1:] {
		if len(rec) < 4 || strings.TrimSpace(rec[0]) == "" {
			continue
		}
		rows = append(rows, CapacityNodeTypeRow{
			NodeType:         strings.TrimSpace(rec[0]),
			Topology:         strings.TrimSpace(rec[1]),
			IsCapacityNode:   parseBoolStr(rec[2]),
			ActiveConstraint: parseBoolStr(rec[3]),
		})
	}
	return rows, nil
}

// ParseDependenciesCSV parses the Dependencies CSV file.
func ParseDependenciesCSV(filePath string) ([]DependencyRuleRow, error) {
	records, err := ReadCSV(filePath)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file: %s", filePath)
	}
	if len(records[0]) < 5 {
		return nil, fmt.Errorf("invalid dependencies CSV header in %s: expected 5 columns, got %d", filePath, len(records[0]))
	}

	var rows []DependencyRuleRow
	for _, rec := range records[1:] {
		if len(rec) < 5 || strings.TrimSpace(rec[0]) == "" || strings.TrimSpace(rec[1]) == "" {
			continue
		}
		rows = append(rows, DependencyRuleRow{
			NodeType:                strings.TrimSpace(rec[0]),
			DependencyNodeType:      strings.TrimSpace(rec[1]),
			RelationshipType:        strings.TrimSpace(rec[2]),
			TopologicalRelationship: strings.TrimSpace(rec[3]),
			UpstreamLevel:           parseOptionalInt(rec[4]),
		})
	}
	return rows, nil
}

// ParseImpactsCSV parses the Impacts CSV file.
func ParseImpactsCSV(filePath string) ([]ImpactRuleRow, error) {
	records, err := ReadCSV(filePath)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file: %s", filePath)
	}
	if len(records[0]) < 4 {
		return nil, fmt.Errorf("invalid impacts CSV header in %s: expected 4 columns, got %d", filePath, len(records[0]))
	}

	var rows []ImpactRuleRow
	for _, rec := range records[1:] {
		if len(rec) < 4 || strings.TrimSpace(rec[0]) == "" || strings.TrimSpace(rec[1]) == "" {
			continue
		}
		rows = append(rows, ImpactRuleRow{
			NodeType:                strings.TrimSpace(rec[0]),
			ImpactNodeType:          strings.TrimSpace(rec[1]),
			TopologicalRelationship: strings.TrimSpace(rec[2]),
			DownstreamLevel:         parseOptionalInt(rec[3]),
		})
	}
	return rows, nil
}

func parseBoolStr(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "true")
}

func parseOptionalInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}
