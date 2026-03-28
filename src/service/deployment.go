package service

import (
	"github.com/killi1812/go-cache-server/app"
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

func (d *DeploymentSrv) Create(workspace, agent string) (any, error) {
	zap.S().Infof("Creating deployment for %s/%s - NOT IMPLEMENTED", workspace, agent)
	return nil, nil
}
