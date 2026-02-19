// package service contains logic for modules
package service

import (
	"errors"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AgentSrv struct {
	db *gorm.DB
}

func NewAgentSrv() *AgentSrv {
	var srv *AgentSrv

	app.Invoke(func(db *gorm.DB) {
		srv = &AgentSrv{db}
	})

	return srv
}

type AgentCreateArgs struct {
	AgentName     string // Required
	WorkspaceName string // Required
	Token         string // Required
}

func (a *AgentSrv) Create(args AgentCreateArgs) (*model.Agent, error) {
	zap.S().Debugf("Adding agent '%s' to workspace '%s'", args.AgentName, args.WorkspaceName)

	zap.S().Debugf("Checking for duplicate agents ")
	var existing model.Agent
	err := a.db.
		Where("name = ?", args.AgentName).
		First(&existing).Error
	if err == nil {
		zap.S().Error("Error Creating new Agent, err: %v", ErrExists)
		return nil, ErrExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.S().Errorf("Error Creating new Agent, err: %v", err)
		return nil, err
	}

	zap.S().Debugf("Trying to retrive workspace %s", args.WorkspaceName)
	// retrive binary workspace
	var workspace model.Workspace
	err = a.db.
		Where("name = ?", args.WorkspaceName).
		First(&workspace).Error
	if err != nil {
		zap.S().Errorf("Failed to retrive workspace %s", args.WorkspaceName)
		return nil, err
	}

	agent := model.Agent{
		Name:        args.AgentName,
		Token:       args.Token,
		WorkspaceId: workspace.ID,
		Workspace:   &workspace,
	}

	if err := a.db.Create(&agent).Error; err != nil {
		zap.S().Errorf("Failed to save agent to database: %v", err)
		return nil, err
	}

	zap.S().Infof("Agent '%s' created successfully (ID: %d)", agent.Name, agent.ID)
	return &agent, nil
}
