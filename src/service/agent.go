// Package service contains logic for modules
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
	zap.S().Infof("Adding agent '%s' to workspace '%s'", args.AgentName, args.WorkspaceName)

	zap.S().Debugf("Checking for duplicate agents ")
	var existing model.Agent
	err := a.db.
		Where("name = ?", args.AgentName).
		First(&existing).Error
	if err == nil {
		zap.S().Errorf("Error Creating new Agent, err: %v", ErrExists)
		return nil, ErrExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.S().Errorf("Error Creating new Agent, err: %v", err)
		return nil, err
	}

	zap.S().Debugf("Trying to retrieve workspace '%s'", args.WorkspaceName)
	// retrieve binary workspace
	var workspace model.Workspace
	err = a.db.
		Where("name = ?", args.WorkspaceName).
		First(&workspace).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve workspace '%s'", args.WorkspaceName)
		return nil, err
	}

	agent := model.Agent{
		Name:        args.AgentName,
		WorkspaceId: workspace.ID,
		Workspace:   &workspace,
		Token:       args.Token,
	}

	if err := a.db.Create(&agent).Error; err != nil {
		zap.S().Errorf("Failed to save agent to database: %v", err)
		return nil, err
	}

	zap.S().Infof("Agent '%s' created successfully (ID: %d)", agent.Name, agent.ID)
	return &agent, nil
}

func (a *AgentSrv) ReadAll(workspace string) ([]model.Agent, error) {
	var agents []model.Agent
	zap.S().Infof("Reading Agents for workspace '%s'", workspace)

	err := a.db.
		InnerJoins("Workspace").
		Where("workspace.name = ?", workspace).
		Find(&agents).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve agents, err: %v", err)
		return nil, err
	}

	return agents, nil
}

func (a *AgentSrv) Read(name string) (*model.Agent, error) {
	var agent model.Agent
	zap.S().Infof("Reading Agent '%s'", name)

	err := a.db.
		Preload("Workspace").
		Where("name = ?", name).
		First(&agent).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve workspace %s, err: %v", name, err)
		return nil, err
	}

	return &agent, nil
}

func (a *AgentSrv) Delete(name string) error {
	zap.S().Warnf("Deleting Agent '%s'", name)

	tx := a.db.Where("name = ?", name).Delete(&model.Agent{})
	if tx.Error != nil {
		return tx.Error
	}

	zap.S().Infof("Agent '%s' removed successfully", name)

	return nil
}
