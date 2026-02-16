package model

import "gorm.io/gorm"

// BinaryCache represents a cache that stores multiple paths and belongs to multiple workspaces.
type BinaryCache struct {
	gorm.Model
	Name      string `gorm:"type:varchar(100);unique;not null" json:"name"`
	URL       string `gorm:"type:varchar(255);not null" json:"url"`
	Token     string `gorm:"type:varchar(255);not null" json:"-"` // not serialized
	Access    string `gorm:"type:varchar(50)" json:"access"`
	Port      int    `json:"port"`
	Retention int    `json:"retention"`

	StorePaths []StorePath `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:CASCADE;" json:"store_paths,omitempty"`
	Workspaces []Workspace `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:SET NULL;" json:"workspaces,omitempty"`
}
