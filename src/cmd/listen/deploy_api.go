package listen

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/service"
)

type deployApi struct {
	agentServ      *service.AgentSrv
	workspaceServ  *service.WorkspaceSrv
	deploymentServ *service.DeploymentSrv
}

func newDeployApi() *deployApi {
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

func (api *deployApi) RegisterEndpoints(group *gin.RouterGroup) {
	// deployment endpoints
	group.GET("/deployment/:workspace", api.getDeployment)
	group.GET("/deployment/:workspace/:name", api.getDeployments)
	group.POST("/deployment/:workspace/:name", api.createDeployment)
	group.GET("/deployment/:workspace/:name/:index", api.getDeploymentByIndex)

	// agent endpoints
	group.GET("/agent/:workspace/:name", api.getAgent)
	group.POST("/agent/:workspace/:name", api.createAgent)
	group.DELETE("/agent/:workspace/:name", api.deleteAgent)
	group.GET("/workspace/:workspace/agents", api.listAgents)

	// workspace endpoints
	group.POST("/workspace", api.createWorkspace)
	group.DELETE("/workspace/:workspace", api.deleteWorkspace)
	group.GET("/workspace/:workspace", api.getWorkspace)
}
