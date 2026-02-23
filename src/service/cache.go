package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrExists = errors.New("err object already exists")

type CacheSrv struct {
	db *gorm.DB
}

func NewCacheSrv() *CacheSrv {
	var srv *CacheSrv

	app.Invoke(func(db *gorm.DB) {
		srv = &CacheSrv{db}
	})

	return srv
}

type CreateCacheArgs struct {
	Name  string // required
	Token string // required
	Port  int    // required

	Retention int // optional default zero
}

// Create handles the GORM logic, token generation, and directory setup for a binary cache.
func (c *CacheSrv) Create(args CreateCacheArgs) (*model.BinaryCache, error) {
	zap.S().Debugf("Creating binary cache '%s' on port %d", args.Name, args.Port)

	var existing model.BinaryCache
	err := c.db.Where("name = ?", args.Name).First(&existing).Error
	if err == nil {
		zap.S().Errorf("Error Creating new cache, err: %v", ErrExists)
		return nil, ErrExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		zap.S().Errorf("Error Creating new cache, err: %v", err)
		return nil, err
	}

	cache := model.BinaryCache{
		Name:      args.Name,
		Uuid:      uuid.New(),
		Port:      args.Port,
		Token:     args.Token,
		Retention: args.Retention,

		URL: fmt.Sprintf("http://%s:%d", config.Config.CacheServer.Hostname, args.Port),
	}

	if err := c.db.Create(&cache).Error; err != nil {
		zap.S().Errorf("Failed to save cache to database: %v", err)
		return nil, err
	}

	zap.S().Infof("Binary cache '%s' created successfully (ID: %d)", cache.Name, cache.ID)
	return &cache, nil
}

func (c *CacheSrv) Delete(name string) error {
	zap.S().Debugf("Removing binary cache %s", name)

	tx := c.db.Where("name = ?", name).Delete(&model.BinaryCache{})
	if tx.Error != nil {
		return tx.Error
	}

	zap.S().Infof("Binary cache %s removed successfully", name)
	return nil
}

func (c *CacheSrv) Read(name string) (*model.BinaryCache, error) {
	var cache model.BinaryCache
	zap.S().Debugf("Reading binary cache %s", name)

	err := c.db.Where("name = ?", name).First(&cache).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve binary cache %s, err: %v", name, err)
		return nil, err
	}

	return &cache, nil
}

func (c *CacheSrv) ReadAll() ([]model.BinaryCache, error) {
	var caches []model.BinaryCache
	zap.S().Debugf("Reading binary cache ")

	err := c.db.Find(&caches).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve multiple binary caches , err: %v", err)
		return nil, err
	}

	return caches, nil
}

func (c *CacheSrv) Update(name string, newCache model.BinaryCache) (*model.BinaryCache, error) {
	var cache *model.BinaryCache
	zap.S().Debugf("Reading binary cache ")

	err := c.db.
		Where("name = ?", name).
		First(&cache).Error
	if err != nil {
		zap.S().Errorf("Failed to retrieve binary cache %s, err: %v", name, err)
		return nil, err
	}

	zap.S().Infof("Old cache %+v", cache)
	if newCache.Name != "" {
		cache.Name = newCache.Name
	}
	if newCache.Port != 0 {
		cache.Port = newCache.Port
		cache.URL = fmt.Sprintf("http://%s:%d", config.Config.CacheServer.Hostname, newCache.Port)
	}
	if newCache.Access != "" {
		cache.Access = newCache.Access
	}
	if newCache.Retention != -1 {
		cache.Retention = newCache.Retention
	}

	zap.S().Infof("New cache %+v", cache)

	err = c.db.Save(cache).Error
	if err != nil {
		zap.S().Errorf("Failed to update cache")
		return nil, err
	}

	return cache, nil
}
