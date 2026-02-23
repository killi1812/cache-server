package cache

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/auth"
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
	router.Use(auth.Protect(s.cache.Token))
	router.GET("/nix-cache-info")
	router.GET("/:storeHash", storeHashCmd)
	router.HEAD("/:storeHash", storeHashCmd)
	// router.GET("/nar/:fileHash") //.nar.:compression

	router.PUT("/:narUuid")
}

var storeHashCmd gin.HandlerFunc = func(c *gin.Context) {
	filename := c.Param("storeHash")

	if before, ok := strings.CutSuffix(filename, ".narinfo"); ok {
		storeHash := before
		// Your logic for .narinfo here
		c.JSON(200, gin.H{"type": "narinfo", "hash": storeHash})
		return
	}

	if before, ok := strings.CutSuffix(filename, ".ls"); ok {
		storeHash := before
		zap.S().Infof("list store hash: '%s'", storeHash)
		// Your logic for .ls here
		c.JSON(200, gin.H{"type": "ls", "hash": storeHash})
		return
	}

	c.String(http.StatusNotFound, "Command not found")
}
