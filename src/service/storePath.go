package service

import (
	"github.com/killi1812/go-cache-server/app"
	"gorm.io/gorm"
)

type StorePathSrv struct {
	db *gorm.DB
}

func NewStorePathSrv() *StorePathSrv {
	var srv *StorePathSrv

	app.Provide(func(db *gorm.DB) {
		srv = &StorePathSrv{db}
	})

	return srv
}
