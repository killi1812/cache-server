package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For now
	},
}

func (api *deployApi) wsHandler(c *gin.Context) {
	name := c.Query("name")
	token := c.Query("token")

	if name == "" || token == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "missing agent name or token"})
		return
	}

	// Validate agent and token
	agent, err := api.agentServ.Read(name)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "agent not found"})
		return
	}

	if agent.Token != token {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.S().Errorf("Failed to upgrade to websocket: %v", err)
		return
	}

	api.hub.Register(name, conn)

	// Keep connection alive and listen for status updates
	go func() {
		defer func() {
			api.hub.Unregister(name)
			conn.Close()
		}()

		for {
			var msg map[string]any
			err := conn.ReadJSON(&msg)
			if err != nil {
				zap.S().Infof("Agent '%s' connection closed: %v", name, err)
				break
			}
			// Handle status updates from agent
			method, _ := msg["method"].(string)
			if method == "DeploymentFinished" {
				command, _ := msg["command"].(map[string]any)
				id, _ := command["id"].(string)
				success, _ := command["hasSucceeded"].(bool)

				status := model.DeploymentSuccess
				if !success {
					status = model.DeploymentFailed
				}

				err := api.deploymentServ.UpdateStatus(id, status)
				if err != nil {
					zap.S().Errorf("Failed to update deployment %s status: %v", id, err)
				} else {
					zap.S().Infof("Deployment %s marked as %s", id, status)
				}
			}
			zap.S().Infof("Received from agent '%s': %v", name, msg)
		}
	}()
}

// getAgent godoc
//
//	@Summary		Get agent info
//	@Description	Get detailed information about an agent in a workspace.
//	@Tags			agent
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Param			name		path		string	true	"Agent Name"
//	@Success		200			{object}	model.Agent
//	@Failure		404			{object}	model.ErrorResponse
//	@Router			/deploy/agent/{workspace}/{name} [get]
func (api *deployApi) getAgent(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Trying to read agent '%s' in workspace '%s'", name, workspace)

	agent, err := api.agentServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read agent '%s', err: %v", name, err)
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "agent not found",
		})
		return
	}

	if agent.Workspace == nil || agent.Workspace.Name != workspace {
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "agent not found in this workspace",
		})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// createAgent godoc
//
//	@Summary		Create agent
//	@Description	Create a new agent in a workspace.
//	@Tags			agent
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Param			name		path		string	true	"Agent Name"
//	@Success		201			{object}	model.Agent
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/deploy/agent/{workspace}/{name} [post]
func (api *deployApi) createAgent(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Trying to create agent '%s' in workspace '%s'", name, workspace)

	t, err := auth.GenerateJwt(workspace)
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to generate token",
		})
		return
	}

	args := service.AgentCreateArgs{
		AgentName:     name,
		WorkspaceName: workspace,
		Token:         t,
	}

	agent, err := api.agentServ.Create(args)
	if err != nil {
		zap.S().Errorf("Failed to create agent, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to create agent",
		})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// deleteAgent godoc
//
//	@Summary		Delete agent
//	@Description	Delete an agent from a workspace.
//	@Tags			agent
//	@Param			workspace	path	string	true	"Workspace Name"
//	@Param			name		path	string	true	"Agent Name"
//	@Success		204
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/deploy/agent/{workspace}/{name} [delete]
func (api *deployApi) deleteAgent(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Trying to delete agent '%s' in workspace '%s'", name, workspace)

	agent, err := api.agentServ.Read(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "agent not found",
		})
		return
	}
	if agent.Workspace == nil || agent.Workspace.Name != workspace {
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "agent not found in this workspace",
		})
		return
	}

	err = api.agentServ.Delete(name)
	if err != nil {
		zap.S().Errorf("Failed to delete agent, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to delete agent",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// listAgents godoc
//
//	@Summary		List agents
//	@Description	List all agents in a workspace.
//	@Tags			workspace
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Success		200			{array}		model.Agent
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/deploy/workspace/{workspace}/agents [get]
func (api *deployApi) listAgents(c *gin.Context) {
	workspace := c.Param("workspace")
	zap.S().Infof("Trying to list agents in workspace '%s'", workspace)

	agents, err := api.agentServ.ReadAll(workspace)
	if err != nil {
		zap.S().Errorf("Failed to read agents, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to list agents",
		})
		return
	}

	c.JSON(http.StatusOK, agents)
}

type WorkspaceRequest struct {
	Name      string `json:"name" binding:"required"`
	CacheName string `json:"cacheName" binding:"required"`
}

// createWorkspace godoc
//
//	@Summary		Create workspace
//	@Description	Create a new workspace associated with a binary cache.
//	@Tags			workspace
//	@Accept			json
//	@Produce		json
//	@Param			request	body		WorkspaceRequest	true	"Workspace details"
//	@Success		201		{object}	model.Workspace
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/deploy/workspace [post]
func (api *deployApi) createWorkspace(c *gin.Context) {
	var req WorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	zap.S().Infof("Trying to create workspace '%s' with cache '%s'", req.Name, req.CacheName)

	t, err := auth.GenerateJwt(req.Name)
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to generate token",
		})
		return
	}

	args := service.WorkspaceCreateArgs{
		WorkspaceName:   req.Name,
		BinaryCacheName: req.CacheName,
		Token:           t,
	}

	workspace, err := api.workspaceServ.Create(args)
	if err != nil {
		zap.S().Errorf("Failed to create workspace, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to create workspace",
		})
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

// deleteWorkspace godoc
//
//	@Summary		Delete workspace
//	@Description	Delete a workspace by name.
//	@Tags			workspace
//	@Param			workspace	path	string	true	"Workspace Name"
//	@Success		204
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/deploy/workspace/{workspace} [delete]
func (api *deployApi) deleteWorkspace(c *gin.Context) {
	name := c.Param("workspace")
	zap.S().Infof("Trying to delete workspace '%s'", name)

	err := api.workspaceServ.Delete(name)
	if err != nil {
		zap.S().Errorf("Failed to delete workspace, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to delete workspace",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// getWorkspace godoc
//
//	@Summary		Get workspace info
//	@Description	Get detailed information about a workspace.
//	@Tags			workspace
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Success		200			{object}	model.Workspace
//	@Failure		404			{object}	model.ErrorResponse
//	@Router			/deploy/workspace/{workspace} [get]
func (api *deployApi) getWorkspace(c *gin.Context) {
	name := c.Param("workspace")
	zap.S().Infof("Trying to read workspace '%s'", name)

	workspace, err := api.workspaceServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read workspace '%s', err: %v", name, err)
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "workspace not found",
		})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// getDeployment godoc
//
//	@Summary		Get deployment info
//	@Description	Get detailed information about a deployment by UUID.
//	@Tags			deployment
//	@Produce		json
//	@Param			workspace	path		string	true	"Deployment UUID"
//	@Success		200			{object}	model.Deployment
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		404			{object}	model.ErrorResponse
//	@Router			/deploy/deployment/{workspace} [get]
func (api *deployApi) getDeployment(c *gin.Context) {
	uuid := c.Param("workspace") // param is ":workspace" in route, but it's used as UUID
	if uuid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "missing deployment UUID",
		})
		return
	}
	zap.S().Infof("Get deployment %s", uuid)
	deployment, err := api.deploymentServ.Read(uuid)
	if err != nil {
		zap.S().Errorf("Failed to read deployment %s, err: %v", uuid, err)
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "deployment not found",
		})
		return
	}
	if deployment == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "deployment not found",
		})
		return
	}
	c.JSON(http.StatusOK, deployment)
}

// getDeployments godoc
//
//	@Summary		Get deployments for agent
//	@Description	List all deployments for a specific agent in a workspace.
//	@Tags			deployment
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Param			name		path		string	true	"Agent Name"
//	@Success		200			{array}		model.Deployment
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/deploy/deployment/{workspace}/{name} [get]
func (api *deployApi) getDeployments(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	if workspace == "" || name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "missing workspace or agent name",
		})
		return
	}
	zap.S().Infof("Get deployments for %s/%s", workspace, name)
	deployments, err := api.deploymentServ.ReadAll(workspace, name)
	if err != nil {
		zap.S().Errorf("Failed to read deployments for %s/%s, err: %v", workspace, name, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to list deployments",
		})
		return
	}
	c.JSON(http.StatusOK, deployments)
}

// createDeployment godoc
//
//	@Summary		Create deployment
//	@Description	Create a new deployment for an agent.
//	@Tags			deployment
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Param			name		path		string	true	"Agent Name"
//	@Success		201			{object}	model.Deployment
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/deploy/deployment/{workspace}/{name} [post]
func (api *deployApi) createDeployment(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	if workspace == "" || name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "missing workspace or agent name",
		})
		return
	}
	zap.S().Infof("Create deployment for %s/%s", workspace, name)
	deployment, err := api.deploymentServ.Create(name, workspace) // Corrected: agentName is 'name'
	if err != nil {
		zap.S().Errorf("Failed to create deployment for %s/%s, err: %v", workspace, name, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to create deployment",
		})
		return
	}
	c.JSON(http.StatusCreated, deployment)
}

// getDeploymentByIndex godoc
//
//	@Summary		Get deployment by index
//	@Description	Get a specific deployment for an agent by its index.
//	@Tags			deployment
//	@Produce		json
//	@Param			workspace	path		string	true	"Workspace Name"
//	@Param			name		path		string	true	"Agent Name"
//	@Param			index		path		string	true	"Deployment Index"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	model.ErrorResponse
//	@Router			/deploy/deployment/{workspace}/{name}/{index} [get]
func (api *deployApi) getDeploymentByIndex(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	index := c.Param("index")
	if workspace == "" || name == "" || index == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "missing workspace, agent name, or index",
		})
		return
	}
	zap.S().Infof("Get deployment %s for %s/%s", index, workspace, name)
	// Just placeholder
	c.JSON(http.StatusOK, gin.H{"index": index, "workspace": workspace, "agent": name})
}

type ActivateRequest struct {
	Agents map[string]string `json:"agents"`
}

// activateDeployment godoc
//
//	@Summary		Activate deployment
//	@Description	Activate deployments for multiple agents.
//	@Tags			deployment
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ActivateRequest	true	"Activation details"
//	@Success		200		{array}		model.Deployment
//	@Failure		400		{object}	model.ErrorResponse
//	@Router			/deploy/activate [post]
func (api *deployApi) activateDeployment(c *gin.Context) {
	var req ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	zap.S().Infof("Activating deployment for agents: %v", req.Agents)

	var deployments []*model.Deployment
	errs := make([]string, 0)
	for agentName, storePath := range req.Agents {
		deployment, err := api.deploymentServ.Create(agentName, storePath)
		if err != nil {
			zap.S().Errorf("Failed to create deployment for agent %s: %v", agentName, err)
			errs = append(errs, err.Error())
			continue
		}

		// Notify agent via WebSocket Hub
		msg := map[string]any{
			"method": "Deployment",
			"command": map[string]any{
				"tag": "Deployment",
				"contents": map[string]any{
					"id":        deployment.Uuid.String(),
					"storePath": storePath,
					"index":     0,
				},
			},
		}
		_ = api.hub.NotifyAgent(agentName, msg)

		deployments = append(deployments, deployment)
	}

	if len(errs) != 0 {
		c.JSON(http.StatusMultiStatus, model.ErrorResponse{
			Error:          "Some deployments failed",
			AdditionalInfo: errs,
		})
		return
	}

	c.JSON(http.StatusOK, deployments)
}
