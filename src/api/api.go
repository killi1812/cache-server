package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	_ "github.com/killi1812/go-cache-server/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

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
	router.GET("/swagger/*any", ginSwagger.CustomWrapHandler(&ginSwagger.Config{URL: "/swagger/doc.json"}, swaggerfiles.Handler))

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
