package cache

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
)

type socketApi struct {
	cache *model.BinaryCache
}

func newApi(cache *model.BinaryCache) app.GinApi {
	return &socketApi{
		cache,
	}
}

// RegisterEndpoints implements app.GinApi.
func (s *socketApi) RegisterEndpoints(router *gin.Engine) {
	router.GET("/nix-cache-info")
	router.GET("/:storeHash.narinfo") // TODO: see how it works with dot
	router.HEAD("/:storeHash.narinfo")
	router.GET("/:storeHash.ls", ls)
	router.GET("/nar/:fileHash.nar.:compression")

	router.PUT("/:narUuid")
}

var ls gin.HandlerFunc = func(c *gin.Context) {
	hash := c.Param("storeHash")

	zap.S().Infof("list store hash: '%s'", hash)

	c.String(200, "Accessing store: %s", hash)
}
