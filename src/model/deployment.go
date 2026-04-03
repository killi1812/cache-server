package model

import (
	"time"

	"github.com/google/uuid"
)

type DeploymentStatus string

const (
	DeploymentPending    DeploymentStatus = "pending"
	DeploymentInProgress DeploymentStatus = "in_progress"
	DeploymentSuccess    DeploymentStatus = "success"
	DeploymentFailed     DeploymentStatus = "failed"
)

type Deployment struct {
	ID        uint             `gorm:"primarykey"`
	Uuid      uuid.UUID        `gorm:"type:uuid;unique;not null" json:"id"`
	CreatedAt time.Time        `json:"createdAt"`
	Status    DeploymentStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	StorePath string           `gorm:"type:varchar(255);not null" json:"storePath"`

	AgentID uint   `json:"agentId"`
	Agent   *Agent `gorm:"foreignKey:AgentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"agent,omitempty"`
}
