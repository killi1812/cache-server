package listen

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"go.uber.org/zap"
)

type mainApi struct {
	cacheServ *service.CacheSrv
}

func newApi() app.GinApi {
	var api *mainApi
	app.Invoke(func(cacheServ *service.CacheSrv) {
		api = &mainApi{cacheServ}
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

func (api *mainApi) name(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to read cache '%s'", name)

	cache, err := api.cacheServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache '%s', err: %v", name, err)
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to read cache"})
		return
	}

	if cache.Access == model.Private {
		// TODO: protect
		// not like this this is middleware only
		auth.Protect(cache.Token)
	}

	/*
		return json.dumps({
		            'githubUsername': '',
		            'isPublic': (self.access == 'public'),
		            'name': self.name,
		            'permission': permission, #TODO
		            'preferredCompressionMethod': 'XZ',
		            'publicSigningKeys': [public_key],
		            'uri': self.url
		        })
	*/
	c.JSON(http.StatusOK, cache)
}
