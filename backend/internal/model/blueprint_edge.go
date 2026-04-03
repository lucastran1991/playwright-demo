package model

import "time"

// BlueprintEdge represents a parent-to-child relationship within a blueprint domain.
type BlueprintEdge struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	BlueprintTypeID uint      `gorm:"uniqueIndex:idx_edge_type_from_to;not null" json:"blueprint_type_id"`
	FromNodeID      uint      `gorm:"uniqueIndex:idx_edge_type_from_to;index;not null" json:"from_node_id"`
	ToNodeID        uint      `gorm:"uniqueIndex:idx_edge_type_from_to;index;not null" json:"to_node_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	BlueprintType BlueprintType `gorm:"foreignKey:BlueprintTypeID" json:"blueprint_type,omitempty"`
	FromNode      BlueprintNode `gorm:"foreignKey:FromNodeID" json:"from_node,omitempty"`
	ToNode        BlueprintNode `gorm:"foreignKey:ToNodeID" json:"to_node,omitempty"`
}
