// package service contains logic for modules
package service

import (
	"github.com/killi1812/go-cache-server/app"
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
