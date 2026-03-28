package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
	"go.uber.org/zap"
)

type deployApi struct {
	agentServ      *service.AgentSrv
	workspaceServ  *service.WorkspaceSrv
	deploymentServ *service.DeploymentSrv
}

func newDeployApi() app.GinApi {
	var api *deployApi
	app.Invoke(func(
		agentServ *service.AgentSrv,
		workspaceServ *service.WorkspaceSrv,
		deploymentServ *service.DeploymentSrv,
	) {
		api = &deployApi{agentServ, workspaceServ, deploymentServ}
	})
	return api
}

func (api *deployApi) RegisterEndpoints(routerGroupByVersion ...*gin.RouterGroup) {
	if len(routerGroupByVersion) == 0 {
		zap.S().Warn("Did not register any endpoints")
		return
	}
	v1 := routerGroupByVersion[0]

	// deployment endpoints
	v1.GET("/deployment/:workspace", api.getDeployment)
	v1.GET("/deployment/:workspace/:name", api.getDeployments)
	v1.POST("/deployment/:workspace/:name", api.createDeployment)
	v1.GET("/deployment/:workspace/:name/:index", api.getDeploymentByIndex)

	// agent endpoints
	v1.GET("/agent/:workspace/:name", api.getAgent)
	v1.POST("/agent/:workspace/:name", api.createAgent)
	v1.DELETE("/agent/:workspace/:name", api.deleteAgent)
	v1.GET("/workspace/:workspace/agents", api.listAgents)

	// workspace endpoints
	v1.POST("/workspace", api.createWorkspace)
	v1.DELETE("/workspace/:workspace", api.deleteWorkspace)
	v1.GET("/workspace/:workspace", api.getWorkspace)

	if len(routerGroupByVersion) == 1 {
		zap.S().Infof("Regester v1 apis")
		return
	}
}
