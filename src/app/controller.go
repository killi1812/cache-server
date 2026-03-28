package app

import (
	"github.com/gin-gonic/gin"
)

type GinApi interface {
	RegisterEndpoints(routerGroupByVersion ...*gin.RouterGroup)
}

type CreateGinApi interface {
	NewGinApi(router *gin.Engine)
}
