package model

import "github.com/google/uuid"

// Agent represents a deployment agent belonging to a workspace.
type Agent struct {
	ID    uint      `gorm:"primarykey" json:"-"`
	Uuid  uuid.UUID `gorm:"type:uuid;unique;not null" json:"id"`
	Name  string    `gorm:"size:255;unique;not null" json:"name"`
	Token string    `gorm:"size:4096;not null" json:"-"`

	WorkspaceId uint       `json:"-"`
	Workspace   *Workspace `gorm:"foreignKey:WorkspaceId;constraint:OnUpdate:CASCADE;" json:"-"`
}
