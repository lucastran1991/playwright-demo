package service

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/user/app/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ModelIngestionSummary holds results of model CSV ingestion.
type ModelIngestionSummary struct {
	CapacityNodesUpserted   int      `json:"capacity_nodes_upserted"`
	DependencyRulesUpserted int      `json:"dependency_rules_upserted"`
	ImpactRulesUpserted     int      `json:"impact_rules_upserted"`
	Errors                  []string `json:"errors,omitempty"`
	DurationMs              int64    `json:"duration_ms"`
}

// ModelIngestionService orchestrates ingestion of model CSVs.
type ModelIngestionService struct {
	db *gorm.DB
}

// NewModelIngestionService creates a new ModelIngestionService.
func NewModelIngestionService(db *gorm.DB) *ModelIngestionService {
	return &ModelIngestionService{db: db}
}

// IngestAll parses and upserts all 3 model CSVs from basePath.
func (s *ModelIngestionService) IngestAll(basePath string) (*ModelIngestionSummary, error) {
	start := time.Now()
	summary := &ModelIngestionSummary{}

	// Parse all 3 CSVs
	capRows, err := ParseCapacityNodesCSV(filepath.Join(basePath, "Capacity Nodes.csv"))
	if err != nil {
		return nil, fmt.Errorf("parse capacity nodes: %w", err)
	}

	depRows, err := ParseDependenciesCSV(filepath.Join(basePath, "Dependencies.csv"))
	if err != nil {
		return nil, fmt.Errorf("parse dependencies: %w", err)
	}

	impRows, err := ParseImpactsCSV(filepath.Join(basePath, "Impacts.csv"))
	if err != nil {
		return nil, fmt.Errorf("parse impacts: %w", err)
	}

	// Upsert in single transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, row := range capRows {
			m := &model.CapacityNodeType{
				NodeType:         row.NodeType,
				Topology:         row.Topology,
				IsCapacityNode:   row.IsCapacityNode,
				ActiveConstraint: row.ActiveConstraint,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "node_type"}},
				DoUpdates: clause.AssignmentColumns([]string{"topology", "is_capacity_node", "active_constraint", "updated_at"}),
			}).Create(m).Error; err != nil {
				return fmt.Errorf("upsert capacity node type %s: %w", row.NodeType, err)
			}
		}
		summary.CapacityNodesUpserted = len(capRows)

		for _, row := range depRows {
			m := &model.DependencyRule{
				NodeType:                row.NodeType,
				DependencyNodeType:      row.DependencyNodeType,
				RelationshipType:        row.RelationshipType,
				TopologicalRelationship: row.TopologicalRelationship,
				UpstreamLevel:           row.UpstreamLevel,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "node_type"}, {Name: "dependency_node_type"}},
				DoUpdates: clause.AssignmentColumns([]string{"relationship_type", "topological_relationship", "upstream_level", "updated_at"}),
			}).Create(m).Error; err != nil {
				return fmt.Errorf("upsert dependency rule %s->%s: %w", row.NodeType, row.DependencyNodeType, err)
			}
		}
		summary.DependencyRulesUpserted = len(depRows)

		for _, row := range impRows {
			m := &model.ImpactRule{
				NodeType:                row.NodeType,
				ImpactNodeType:          row.ImpactNodeType,
				TopologicalRelationship: row.TopologicalRelationship,
				DownstreamLevel:         row.DownstreamLevel,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "node_type"}, {Name: "impact_node_type"}},
				DoUpdates: clause.AssignmentColumns([]string{"topological_relationship", "downstream_level", "updated_at"}),
			}).Create(m).Error; err != nil {
				return fmt.Errorf("upsert impact rule %s->%s: %w", row.NodeType, row.ImpactNodeType, err)
			}
		}
		summary.ImpactRulesUpserted = len(impRows)

		return nil
	})

	summary.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		return nil, err
	}
	return summary, nil
}
