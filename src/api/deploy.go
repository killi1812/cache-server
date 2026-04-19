package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"go.uber.org/zap"
)

// deployApi handles RESTful management.
type deployApi struct {
	agentServ      *service.AgentSrv
	workspaceServ  *service.WorkspaceSrv
	deploymentServ *service.DeploymentSrv
	hub            *service.Hub
}

func newDeployApi(
	agentServ *service.AgentSrv,
	workspaceServ *service.WorkspaceSrv,
	deploymentServ *service.DeploymentSrv,
	hub *service.Hub,
) app.GinApi {
	return &deployApi{agentServ, workspaceServ, deploymentServ, hub}
}

func (api *deployApi) RegisterEndpoints(routerGroupByVersion ...*gin.RouterGroup) {
	if len(routerGroupByVersion) == 0 {
		return
	}
	v1 := routerGroupByVersion[0]
	deploy := v1.Group("/deploy")

	// deployment endpoints
	deploy.GET("/deployment/:workspace", api.getDeployment)
	deploy.GET("/deployment/:workspace/:name", api.getDeployments)
	deploy.POST("/deployment/:workspace/:name", api.createDeployment)
	deploy.GET("/deployment/:workspace/:name/:index", api.getDeploymentByIndex)

	// agent endpoints
	deploy.GET("/agent/:workspace/:name", api.getAgent)
	deploy.POST("/agent/:workspace/:name", api.createAgent)
	deploy.DELETE("/agent/:workspace/:name", api.deleteAgent)
	deploy.GET("/workspace/:workspace/agents", api.listAgents)

	// workspace endpoints
	deploy.POST("/workspace", api.createWorkspace)
	deploy.DELETE("/workspace/:workspace", api.deleteWorkspace)
	deploy.GET("/workspace/:workspace", api.getWorkspace)

	deploy.POST("/activate", api.activateDeployment)
}

// deployWsApi handles WebSockets on deploy port.
type deployWsApi struct {
	agentServ      *service.AgentSrv
	deploymentServ *service.DeploymentSrv
	hub            *service.Hub
}

func NewDeployWsApi(
	agentServ *service.AgentSrv,
	deploymentServ *service.DeploymentSrv,
	hub *service.Hub,
) app.CreateGinApi {
	return &deployWsApi{agentServ, deploymentServ, hub}
}

func (api *deployWsApi) NewGinApi(router *gin.Engine) {
	zap.S().Info("Registering WebSocket Deploy API")
	router.GET("/ws", api.wsHandler)
	router.GET("/ws-deployment", api.deploymentHandler)
	router.GET("/api/v1/deploy/log/", api.logHandler)
}
