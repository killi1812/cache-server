package model

import "gorm.io/gorm"

// Workspace represents a deployment environment linked to a specific binary cache.
type Workspace struct {
	gorm.Model
	Name  string `gorm:"type:varchar(100);unique;not null" json:"name"`
	Token string `gorm:"type:varchar(255);not null" json:"token,omitempty"`

	BinaryCacheId uint         `json:"binary_cache_id"`
	BinaryCache   *BinaryCache `gorm:"foreignKey:BinaryCacheId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"binary_cache,omitempty"`

	Agents []Agent `gorm:"foreignKey:WorkspaceId;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"agents,omitempty"`
}
