package service

import (
	"time"

	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/objstor"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GCSrv struct {
	db      *gorm.DB
	storage objstor.ObjectStorage
}

func NewGCSrv(db *gorm.DB, storage objstor.ObjectStorage) *GCSrv {
	return &GCSrv{db, storage}
}

// Start runs the collector every hour.
func (g *GCSrv) Start() {
	zap.S().Info("Starting Garbage Collector worker")
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			g.Collect()
		}
	}()
}

// Collect finds and removes old store paths based on cache retention.
func (g *GCSrv) Collect() {
	zap.S().Info("Garbage Collector: Starting collection cycle")

	var caches []model.BinaryCache
	if err := g.db.Find(&caches).Error; err != nil {
		zap.S().Errorf("GC: Failed to fetch caches: %v", err)
		return
	}

	for _, cache := range caches {
		if cache.Retention <= 0 {
			continue
		}

		// Calculate expiration time (retention in days)
		expiration := time.Now().AddDate(0, 0, -cache.Retention)

		var oldPaths []model.StorePath
		err := g.db.Where("binary_cache_id = ? AND created_at < ?", cache.ID, expiration).Find(&oldPaths).Error
		if err != nil {
			zap.S().Errorf("GC: Failed to fetch old paths for cache '%s': %v", cache.Name, err)
			continue
		}

		if len(oldPaths) == 0 {
			continue
		}

		zap.S().Infof("GC: Removing %d old paths from cache '%s'", len(oldPaths), cache.Name)

		for _, path := range oldPaths {
			// 1. Remove from storage
			err := g.storage.DeleteFile(cache.Name, path.FileHash)
			if err != nil {
				zap.S().Warnf("GC: Failed to delete file '%s' from storage: %v", path.FileHash, err)
				// Continue to next path, don't delete from DB if storage fail?
				// Usually better to keep DB in sync with storage.
			}

			// 2. Permanent delete from DB (Unscoped)
			if err := g.db.Unscoped().Delete(&path).Error; err != nil {
				zap.S().Errorf("GC: Failed to delete DB entry for path '%s': %v", path.StoreHash, err)
			}
		}
	}

	zap.S().Info("Garbage Collector: Collection cycle complete")
}
