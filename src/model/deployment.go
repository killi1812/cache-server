package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeploymentStatus string

const (
	DeploymentPending    DeploymentStatus = "Pending"
	DeploymentInProgress DeploymentStatus = "InProgress"
	DeploymentSuccess    DeploymentStatus = "Success"
	DeploymentFailed     DeploymentStatus = "Failed"
)

type Deployment struct {
	ID        uint             `gorm:"primarykey" json:"-"`
	CreatedAt time.Time        `json:"createdOn"`
	UpdatedAt time.Time        `json:"-"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
	Uuid      uuid.UUID        `gorm:"type:uuid;unique;not null" json:"id"`
	Index     int              `json:"index"`
	Status    DeploymentStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	StorePath string           `gorm:"type:varchar(255);not null" json:"storePath"`

	AgentID uint   `json:"-"`
	Agent   *Agent `gorm:"foreignKey:AgentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"agent,omitempty"`
}
