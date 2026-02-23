package listen

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"go.uber.org/zap"
)

type mainApi struct{}

func newApi() app.GinApi {
	return &mainApi{}
}

// RegisterEndpoints implements app.GinApi.
func (m mainApi) RegisterEndpoints(router *gin.Engine) {
	apiGroup := router.Group("/api")

	// v1 group
	v1 := apiGroup.Group("/v1")
	// cache group
	cache := v1.Group("/cache")
	cache.GET("/:name", name)
	cache.POST("/:name/narinfo")

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

var name gin.HandlerFunc = func(c *gin.Context) {
	hash := c.Param("name")

	zap.S().Infof("list store hash: '%s'", hash)

	c.String(200, "Accessing store: %s", hash)
}
