package service

import (
	"github.com/killi1812/go-cache-server/app"
	"gorm.io/gorm"
)

type WorkspaceSrv struct {
	db *gorm.DB
}

func NewWorkspaceSrv() *WorkspaceSrv {
	var srv *WorkspaceSrv

	app.Provide(func(db *gorm.DB) {
		srv = &WorkspaceSrv{db}
	})

	return srv
}
