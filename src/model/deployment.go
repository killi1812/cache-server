package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeploymentStatus string

const (
	DeploymentPending    DeploymentStatus = "pending"
	DeploymentInProgress DeploymentStatus = "in_progress"
	DeploymentSuccess    DeploymentStatus = "success"
	DeploymentFailed     DeploymentStatus = "failed"
)

type Deployment struct {
	gorm.Model
	Uuid      uuid.UUID        `gorm:"type:uuid;unique;not null" json:"id"`
	Status    DeploymentStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	StorePath string           `gorm:"type:varchar(255);not null" json:"storePath"`

	AgentID uint   `json:"agentId"`
	Agent   *Agent `gorm:"foreignKey:AgentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"agent,omitempty"`
}
