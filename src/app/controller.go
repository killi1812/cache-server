package app

import (
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
