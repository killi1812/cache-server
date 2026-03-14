package listen

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/auth"
	"go.uber.org/zap"
)

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
		auth.Protect(cache.Token)(c)
		if c.IsAborted() {
			return
		}
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

	// TODO: Missing publicSigningKeys from output
	c.JSON(http.StatusOK, cache)
}

func (api *mainApi) narinfo(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to retrive missing narinfo '%s'", name)

	var incomingHashes []string

	if err := c.ShouldBindJSON(&incomingHashes); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	missing, err := api.pathServ.GetMissingHashes(name, incomingHashes)
	if err != nil {
		zap.S().Errorf("Failed to query missing hashes: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, missing)
}
