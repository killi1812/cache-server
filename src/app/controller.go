package app

import (
	"github.com/gin-gonic/gin"
)

type GinApi interface {
	RegisterEndpoints(router *gin.Engine)
}
