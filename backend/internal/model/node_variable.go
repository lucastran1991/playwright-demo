package model

import "time"

// NodeVariable stores per-node capacity metrics as key-value pairs.
// Sparse data (35 CSV columns, most nodes use 3-5) makes KV more efficient than wide table.
type NodeVariable struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	NodeID       string    `gorm:"uniqueIndex:idx_nv_node_var;size:255;not null" json:"node_id"`
	VariableName string    `gorm:"uniqueIndex:idx_nv_node_var;size:100;not null" json:"variable_name"`
	Value        float64   `gorm:"not null" json:"value"`
	Unit         string    `gorm:"size:20;default:'kW'" json:"unit"`
	Source       string    `gorm:"size:20;not null;index" json:"source"` // csv_import | computed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
