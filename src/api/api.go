package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	_ "github.com/killi1812/go-cache-server/docs/management"
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
	deployApi app.GinApi
	cacheApi  app.GinApi
}

func NewApi() app.CreateGinApi {
	var api *Api = &Api{cacheApi: newCacheApi(), deployApi: newDeployApi()}
	return api
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
	api.cacheApi.RegisterEndpoints(v1, v2)

	// deploy group
	api.deployApi.RegisterEndpoints(v1)
}
