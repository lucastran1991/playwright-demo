package model

import "time"

// BlueprintType represents a blueprint domain category (e.g. Cooling, Electrical).
type BlueprintType struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Slug       string    `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	FolderName string    `gorm:"size:255;not null" json:"folder_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
