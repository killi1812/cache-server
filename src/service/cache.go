package service

import (
	"errors"
	"fmt"

	"github.com/killi1812/go-cache-server/app"
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
func (m *CacheSrv) Create(args CreateCacheArgs) (*model.BinaryCache, error) {
	zap.S().Debugf("Database: Creating binary cache '%s' on port %d", args.Name, args.Port)

	var existing model.BinaryCache
	err := m.db.Where("name = ?", args.Name).First(&existing).Error
	if err == nil {
		return nil, errors.Join(ErrExists, err)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cache := model.BinaryCache{
		Name:      args.Name,
		Port:      args.Port,
		Token:     args.Token,
		Retention: args.Retention,

		// TODO: hostname
		URL: fmt.Sprintf("http://localhost:%d", args.Port), // Default URL logic
	}

	if err := m.db.Create(&cache).Error; err != nil {
		return nil, fmt.Errorf("failed to save cache to database: %w", err)
	}

	return &cache, nil
}
