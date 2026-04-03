package model

import "time"

// ImpactRule defines a type-level impact relationship.
// Example: "UPS" impacts "RPP" (Downstream, level 2).
type ImpactRule struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	NodeType                string    `gorm:"uniqueIndex:idx_impact_rule_type_pair;index;size:100;not null" json:"node_type"`
	ImpactNodeType          string    `gorm:"uniqueIndex:idx_impact_rule_type_pair;size:100;not null" json:"impact_node_type"`
	TopologicalRelationship string    `gorm:"size:20;not null" json:"topological_relationship"`
	DownstreamLevel         *int      `json:"downstream_level"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
