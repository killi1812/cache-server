package api

import (
	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
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
	deploy := v1.Group("/deploy")
	deploy.Use(auth.Protect(config.Config.CacheServer.Key))

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

	if len(routerGroupByVersion) == 1 {
		zap.S().Infof("Regester v1 apis")
		return
	}
}
