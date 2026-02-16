// Package service containse a database access
package service

import "github.com/killi1812/go-cache-server/app"

func init() {
	app.Provide(NewAgentSrv)
	app.Provide(NewCacheSrv)
	app.Provide(NewStorePathSrv)
	app.Provide(NewWorkspaceSrv)
}
