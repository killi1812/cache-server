package listen

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
)

type mainApi struct {
	cacheServ *service.CacheSrv
	pathServ  *service.StorePathSrv
}

func newApi() app.GinApi {
	var api *mainApi
	app.Invoke(func(cacheServ *service.CacheSrv, pathServ *service.StorePathSrv) {
		api = &mainApi{cacheServ, pathServ}
	})
	return api
}

// RegisterEndpoints implements app.GinApi.
func (api mainApi) RegisterEndpoints(router *gin.Engine) {
	apiGroup := router.Group("/api")

	// v1 group
	v1 := apiGroup.Group("/v1")
	// cache group
	cache := v1.Group("/cache")
	cache.GET("/:name", api.name)
	cache.POST("/:name/narinfo", api.narinfo)

	cache.POST("/:name/multipart-nar")
	cache.POST("/:name/multipart-nar/:narUuid")
	cache.POST("/:name/multipart-nar/:narUuid/complete")
	cache.POST("/:name/multipart-nar/:narUuid/abort")

	// deploy group
	deploy := v1.Group("/deploy")
	deploy.GET("/deployment/:uuid")

	// v2 group
	v2 := apiGroup.Group("/v2")
	// deploy group
	deployV2 := v2.Group("/deploy")
	deployV2.POST("activate")
}
