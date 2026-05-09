package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeploymentSrv struct {
	db *gorm.DB
}

func NewDeploymentSrv(db *gorm.DB) *DeploymentSrv {
	return &DeploymentSrv{db}
}

func (d *DeploymentSrv) Read(uuidStr string) (*model.Deployment, error) {
	var dep model.Deployment
	err := d.db.Where("uuid = ?", uuidStr).Preload("Agent").First(&dep).Error
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (d *DeploymentSrv) ReadAll(workspace, agent string) ([]model.Deployment, error) {
	var deps []model.Deployment
	err := d.db.Joins("Agent").
		Joins("Agent.Workspace").
		Where("\"Agent__Workspace\".\"name\" = ? AND \"Agent\".\"name\" = ?", workspace, agent).
		Find(&deps).Error
	return deps, err
}

func (d *DeploymentSrv) Create(agentName, storePath string) (*model.Deployment, error) {
	zap.S().Infof("Creating deployment for agent %s: %s", agentName, storePath)

	var agent model.Agent
	err := d.db.Where("name = ?", agentName).First(&agent).Error
	if err != nil {
		return nil, err
	}

	deployment := model.Deployment{
		Uuid:      uuid.New(),
		Status:    model.DeploymentPending,
		StorePath: storePath,
		AgentID:   agent.ID,
	}

	if err := d.db.Create(&deployment).Error; err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (d *DeploymentSrv) UpdateStatus(uuidStr string, status model.DeploymentStatus) error {
	u, err := uuid.Parse(uuidStr)
	if err != nil {
		return fmt.Errorf("invalid uuid format: %w", err)
	}
	result := d.db.Model(&model.Deployment{}).Where("uuid = ?", u).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no deployment found with uuid %s", uuidStr)
	}
	return nil
}
