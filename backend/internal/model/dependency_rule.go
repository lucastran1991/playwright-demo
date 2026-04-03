package model

import "time"

// DependencyRule defines a type-level dependency relationship.
// Example: "Rack" depends on "RPP" (Upstream, level 1).
type DependencyRule struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	NodeType                string    `gorm:"uniqueIndex:idx_dep_rule_type_pair;index;size:100;not null" json:"node_type"`
	DependencyNodeType      string    `gorm:"uniqueIndex:idx_dep_rule_type_pair;size:100;not null" json:"dependency_node_type"`
	RelationshipType        string    `gorm:"size:20;not null" json:"relationship_type"`
	TopologicalRelationship string    `gorm:"size:20;not null" json:"topological_relationship"`
	UpstreamLevel           *int      `json:"upstream_level"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
