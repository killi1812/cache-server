package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BinaryCacheAccess string

const (
	Private BinaryCacheAccess = "private"
	Public  BinaryCacheAccess = "public"
)

// ParseBinaryCacheAccess parses the input, default private
func ParseBinaryCacheAccess(str string) BinaryCacheAccess {
	switch str {

	case "Private":
	case "private":
		return Private

	case "Public":
	case "public":
		return Public
	}

	return Private
}

// BinaryCache represents a cache that stores multiple paths and belongs to multiple workspaces.
type BinaryCache struct {
	gorm.Model
	Uuid      uuid.UUID         `gorm:"type:uuid;unique;not null"`
	Name      string            `gorm:"type:varchar(100);unique;not null" json:"name"`
	URL       string            `gorm:"type:varchar(255);not null" json:"url"`
	Token     string            `gorm:"type:varchar(255);not null" json:"token,omitempty"`
	Access    BinaryCacheAccess `gorm:"type:varchar(50)" json:"access"`
	Port      int               `gorm:"unique" json:"port"`
	Retention int               `json:"retention"`
	PublicKey string            `gorm:"type:varchar(255)" json:"publicKey"`
	SecretKey string            `gorm:"type:varchar(255)" json:"-"` // Not serialized

	StorePaths []StorePath `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:CASCADE;" json:"store_paths,omitempty"`
	Workspaces []Workspace `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:SET NULL;" json:"workspaces,omitempty"`
}
