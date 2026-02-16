package service

import (
	"github.com/killi1812/go-cache-server/app"
	"gorm.io/gorm"
)

type CacheSrv struct {
	db *gorm.DB
}

func NewCacheSrv() *CacheSrv {
	var srv *CacheSrv

	app.Provide(func(db *gorm.DB) {
		srv = &CacheSrv{db}
	})

	return srv
}
