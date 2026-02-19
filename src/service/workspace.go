package service

import (
	"errors"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkspaceSrv struct {
	db *gorm.DB
}

func NewWorkspaceSrv() *WorkspaceSrv {
	var srv *WorkspaceSrv

	app.Invoke(func(db *gorm.DB) {
		srv = &WorkspaceSrv{db}
	})

	return srv
}

type WorkspaceCreateArgs struct {
	BinaryCacheName string // Required
	WorkspaceName   string // Required
	Token           string // Required
}

func (w WorkspaceSrv) Create(args WorkspaceCreateArgs) (*model.Workspace, error) {
	zap.S().Debugf("Creating Workspace %+v", args)

	zap.S().Debugf("Checking for duplicate workspace %s", args.WorkspaceName)
	// check if Workspace already exists
	var existing model.Workspace
	err := w.db.Where("name = ?", args.WorkspaceName).First(&existing).Error
	if err == nil {
		zap.S().Error("Error Creating new Workspace, err: %v", ErrExists)
		return nil, ErrExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.S().Errorf("Error Creating new Workspace, err: %v", err)
		return nil, err
	}

	zap.S().Debugf("Trying to retrive binary cache %s", args.BinaryCacheName)
	// retrive binary cache
	var cache model.BinaryCache
	err = w.db.Where("name = ?", args.BinaryCacheName).First(&cache).Error
	if err != nil {
		zap.S().Errorf("Failed to retrive Cache %s", args.BinaryCacheName)
		return nil, err
	}

	workspace := model.Workspace{
		Name:          args.WorkspaceName,
		Token:         args.Token,
		BinaryCacheId: cache.ID,
		BinaryCache:   &cache,
	}

	if err := w.db.Create(&workspace).Error; err != nil {
		zap.S().Errorf("Failed to save workspace to database: %v", err)
		return nil, err
	}

	zap.S().Infof("Workspace '%s' created successfully (ID: %d)", workspace.Name, workspace.ID)

	return &workspace, nil
}

func (w *WorkspaceSrv) ReadAll() ([]model.Workspace, error) {
	var workspaces []model.Workspace
	zap.S().Debugf("Reading binary cache ")

	err := w.db.Find(&workspaces).Error
	if err != nil {
		zap.S().Errorf("Failed to retrive binary multiple caches , err: %v", err)
		return nil, err
	}

	return workspaces, nil
}
