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
	Name      string            `gorm:"type:text;unique;not null" json:"name"`
	URL       string            `gorm:"type:text;not null" json:"url"`
	Token     string            `gorm:"type:text;not null" json:"token,omitempty"`
	Access    BinaryCacheAccess `gorm:"type:text" json:"access"`
	Port      int               `gorm:"unique" json:"port"`
	Retention int               `json:"retention"`
	PublicKey string            `gorm:"type:text" json:"publicKey"`
	SecretKey string            `gorm:"type:text" json:"-"`

	StorePaths []StorePath `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:CASCADE;" json:"store_paths,omitempty"`
	Workspaces []Workspace `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:SET NULL;" json:"workspaces,omitempty"`
}
