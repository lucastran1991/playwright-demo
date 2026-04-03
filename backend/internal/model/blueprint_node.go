package model

import "time"

// BlueprintNode represents a unified node across all blueprint domains.
// Same physical asset (e.g. a rack) exists as one row regardless of how many domains reference it.
type BlueprintNode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    string    `gorm:"uniqueIndex;size:255;not null" json:"node_id"`
	Name      string    `gorm:"size:500;not null" json:"name"`
	NodeType  string    `gorm:"index;size:100" json:"node_type"`
	NodeRole  string    `gorm:"size:100" json:"node_role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
