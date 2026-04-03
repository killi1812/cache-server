package service

import (
	"github.com/google/uuid"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeploymentSrv struct {
	db *gorm.DB
}

func NewDeploymentSrv() *DeploymentSrv {
	var srv *DeploymentSrv

	app.Invoke(func(db *gorm.DB) {
		srv = &DeploymentSrv{db}
	})

	return srv
}

func (d *DeploymentSrv) Read(uuid string) (any, error) {
	zap.S().Infof("Reading deployment %s - NOT IMPLEMENTED", uuid)
	return nil, nil
}

func (d *DeploymentSrv) ReadAll(workspace, agent string) ([]any, error) {
	zap.S().Infof("Reading deployments for %s/%s - NOT IMPLEMENTED", workspace, agent)
	return nil, nil
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
