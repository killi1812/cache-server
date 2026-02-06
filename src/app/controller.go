package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller interface {
	RegisterEndpoints(router *gin.RouterGroup)
}

var controllers []Controller

// RegisterController registers a controller to a router
func RegisterController(newCtn func() Controller) {
	zap.S().DPanicf("Registering controller %T", newCtn)
	controllers = append(controllers, newCtn())
}
