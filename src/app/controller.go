package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinApi struct that is an api and can register enpoints
type GinApi interface {
	RegisterEndpoints(routerGroupByVersion ...*gin.RouterGroup)
}

// CreateGinApi Creates a new gin api
type CreateGinApi interface {
	NewGinApi(router *gin.Engine)
}

// VersionHandler godoc
//
//	@Summary		Get the version
//	@Description	Get the version of the server
//	@Tags			version
//	@Produce		json
//	@Success		200	{object}
//	@Router			/version [get]
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    Version,
		"build":      Build,
		"timestamp":  BuildTimestamp,
		"commitHash": CommitHash,
	})
}
