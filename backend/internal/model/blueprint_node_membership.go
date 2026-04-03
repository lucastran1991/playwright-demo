package model

import "time"

// BlueprintNodeMembership maps a node to a blueprint domain with domain-specific org path.
type BlueprintNodeMembership struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	BlueprintTypeID uint      `gorm:"uniqueIndex:idx_membership_type_node;not null" json:"blueprint_type_id"`
	BlueprintNodeID uint      `gorm:"uniqueIndex:idx_membership_type_node;not null" json:"blueprint_node_id"`
	OrgPath         string    `gorm:"size:1000" json:"org_path"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	BlueprintType BlueprintType `gorm:"foreignKey:BlueprintTypeID" json:"blueprint_type,omitempty"`
	BlueprintNode BlueprintNode `gorm:"foreignKey:BlueprintNodeID" json:"blueprint_node,omitempty"`
}
