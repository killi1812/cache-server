package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/objstor"
	"go.uber.org/zap"
)

type cacheApi struct {
	cacheServ *service.CacheSrv
	pathServ  *service.StorePathSrv
	storage   objstor.ObjectStorage
}

func newCacheApi() app.GinApi {
	var api *cacheApi
	app.Invoke(func(
		cacheServ *service.CacheSrv,
		pathServ *service.StorePathSrv,
		storage objstor.ObjectStorage,
	) {
		api = &cacheApi{cacheServ, pathServ, storage}
	})
	return api
}

// RegisterEndpoints implements app.GinApi.
func (api *cacheApi) RegisterEndpoints(routerGroupByVersion ...*gin.RouterGroup) {
	if len(routerGroupByVersion) == 0 {
		zap.S().Warn("Did not register any endpoints")
		return
	}
	v1 := routerGroupByVersion[0]

	// cache group
	cache := v1.Group("/cache")
	cache.GET("/:name", api.name)
	cache.POST("/:name/narinfo", api.narinfo)

	cache.POST("/:name/multipart-nar", api.createNar)
	cache.POST("/:name/multipart-nar/:narUuid", api.redirect)
	cache.POST("/:name/multipart-nar/:narUuid/complete")
	cache.POST("/:name/multipart-nar/:narUuid/abort")

	if len(routerGroupByVersion) == 1 {
		zap.S().Infof("Regester v1 apis")
		return
	}
	v2 := routerGroupByVersion[1]
	// deploy group
	deployV2 := v2.Group("/deploy")
	deployV2.POST("activate")

	if len(routerGroupByVersion) == 2 {
		zap.S().Infof("Regester v2 apis")
		return
	}
}
