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
	ws *gorm.DB
}

func NewWorkspaceSrv() *WorkspaceSrv {
	var srv *WorkspaceSrv

	app.Invoke(func(db *gorm.DB) {
		srv = &WorkspaceSrv{
			db: db,
			ws: db.
				Model(&model.Workspace{}).
				Preload("BinaryCache").
				Preload("Agents"),
		}
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

	zap.S().Debugf("Trying to retrieve binary cache %s", args.BinaryCacheName)
	// retrieve binary cache
	var cache model.BinaryCache
	err = w.db.Where("name = ?", args.BinaryCacheName).First(&cache).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve Cache %s", args.BinaryCacheName)
		return nil, err
	}

	workspace := model.Workspace{
		Name:          args.WorkspaceName,
		BinaryCacheId: cache.ID,
		BinaryCache:   &cache,
		Token:         args.Token,
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
	zap.S().Infof("Reading Workspaces")

	err := w.ws.Find(&workspaces).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve multiple workspaces, err: %v", err)
		return nil, err
	}

	return workspaces, nil
}

func (w *WorkspaceSrv) Read(name string) (*model.Workspace, error) {
	var workspace model.Workspace
	zap.S().Infof("Reading workspace %s", name)

	err := w.ws.
		Where("name = ?", name).
		First(&workspace).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve workspace %s, err: %v", name, err)
		return nil, err
	}

	return &workspace, nil
}

func (w *WorkspaceSrv) Delete(name string) error {
	zap.S().Warnf("Deleting workspace '%s' and all associated agents", name)

	tx := w.db.Where("name = ?", name).Delete(&model.Workspace{})
	if tx.Error != nil {
		return tx.Error
	}

	zap.S().Infof("Workspace '%s' removed successfully", name)
	return nil
}

func (w *WorkspaceSrv) UpdateCache(wsName string, cacheName string) (*model.Workspace, error) {
	zap.S().Infof("Updating workspace '%s' to use cache '%s'", wsName, cacheName)

	var cache model.BinaryCache
	err := w.db.
		Where("name = ?", cacheName).
		First(&cache).Error
	if err != nil {
		zap.S().Errorf("Failed to find binary cache %s, err: %v", cacheName, err)
		return nil, err
	}

	err = w.ws.
		Where("name = ?", wsName).
		Update("BinaryCacheId", cache.ID).
		Error
	if err != nil {
		zap.S().Errorf("Failed to update workspace %s, err: %v", wsName, err)
		return nil, err
	}

	var workspace model.Workspace
	err = w.ws.
		Where("name = ?", wsName).
		First(&workspace).Error
	if err != nil {
		zap.S().Errorf("Failed to find updated workspace %s, err: %v", wsName, err)
		return nil, err
	}

	zap.S().Infof("Workspace '%s' updated successfully", workspace.Name)
	return &workspace, nil
}
