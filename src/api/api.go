package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	_ "github.com/killi1812/go-cache-server/docs/management"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/objstor"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Management API
//	@version		1.0
//	@description	API for managing binary caches, workspaces, and agents.
//	@BasePath		/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

type Api struct {
	deployMgmtApi app.GinApi
	cacheMgmtApi  app.GinApi
}

func NewApi(
	cacheServ *service.CacheSrv,
	pathServ *service.StorePathSrv,
	agentServ *service.AgentSrv,
	workspaceServ *service.WorkspaceSrv,
	deploymentServ *service.DeploymentSrv,
	hub *service.Hub,
	storage objstor.ObjectStorage,
) app.CreateGinApi {
	cacheMgmtApi := newCacheApi(cacheServ, pathServ, storage)
	deployMgmtApi := newDeployApi(agentServ, workspaceServ, deploymentServ, hub)
	return &Api{cacheMgmtApi: cacheMgmtApi, deployMgmtApi: deployMgmtApi}
}

// RegisterEndpoints implements app.GinApi.
func (api *Api) NewGinApi(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.CustomWrapHandler(&ginSwagger.Config{
		InstanceName: "management",
		URL:          "doc.json",
	}, swaggerfiles.Handler))

	apiGroup := router.Group("/api")

	// v1 group
	v1 := apiGroup.Group("/v1")
	// v2 group
	v2 := apiGroup.Group("/v2")

	// cache group
	api.cacheMgmtApi.RegisterEndpoints(v1, v2)

	// deploy group
	api.deployMgmtApi.RegisterEndpoints(v1)
}
