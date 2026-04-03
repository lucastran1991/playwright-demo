package model

import "time"

// CapacityNodeType classifies a node type as a capacity domain and/or active constraint.
// This is type-level metadata (24 rows), not per-instance data.
type CapacityNodeType struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	NodeType         string    `gorm:"uniqueIndex;size:100;not null" json:"node_type"`
	Topology         string    `gorm:"size:100;not null" json:"topology"`
	IsCapacityNode   bool      `gorm:"not null;default:false" json:"is_capacity_node"`
	ActiveConstraint bool      `gorm:"not null;default:false" json:"active_constraint"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
