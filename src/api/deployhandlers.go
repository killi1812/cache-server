package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"go.uber.org/zap"
)

func (api *deployApi) getAgent(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Trying to read agent '%s' in workspace '%s'", name, workspace)

	agent, err := api.agentServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read agent '%s', err: %v", name, err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	if agent.Workspace == nil || agent.Workspace.Name != workspace {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "agent not found in this workspace"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (api *deployApi) createAgent(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Trying to create agent '%s' in workspace '%s'", name, workspace)

	t, err := auth.GenerateJwt(workspace)
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

func (api *deployApi) deleteAgent(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Trying to delete agent '%s' in workspace '%s'", name, workspace)

	agent, err := api.agentServ.Read(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	if agent.Workspace == nil || agent.Workspace.Name != workspace {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "agent not found in this workspace"})
		return
	}

	err = api.agentServ.Delete(name)
	if err != nil {
		zap.S().Errorf("Failed to delete agent, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete agent"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (api *deployApi) listAgents(c *gin.Context) {
	workspace := c.Param("workspace")
	zap.S().Infof("Trying to list agents in workspace '%s'", workspace)

	agents, err := api.agentServ.ReadAll(workspace)
	if err != nil {
		zap.S().Errorf("Failed to read agents, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to list agents"})
		return
	}

	c.JSON(http.StatusOK, agents)
}

type WorkspaceRequest struct {
	Name      string `json:"name" binding:"required"`
	CacheName string `json:"cacheName" binding:"required"`
}

func (api *deployApi) createWorkspace(c *gin.Context) {
	var req WorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	zap.S().Infof("Trying to create workspace '%s' with cache '%s'", req.Name, req.CacheName)

	t, err := auth.GenerateJwt(req.Name)
	if err != nil {
		zap.S().Errorf("Failed to generate token, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace"})
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (api *deployApi) deleteWorkspace(c *gin.Context) {
	name := c.Param("workspace")
	zap.S().Infof("Trying to delete workspace '%s'", name)

	err := api.workspaceServ.Delete(name)
	if err != nil {
		zap.S().Errorf("Failed to delete workspace, err: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (api *deployApi) getWorkspace(c *gin.Context) {
	name := c.Param("workspace")
	zap.S().Infof("Trying to read workspace '%s'", name)

	workspace, err := api.workspaceServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read workspace '%s', err: %v", name, err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (api *deployApi) getDeployment(c *gin.Context) {
	uuid := c.Param("workspace")
	zap.S().Infof("Get deployment %s", uuid)
	deployment, _ := api.deploymentServ.Read(uuid)
	c.JSON(http.StatusOK, deployment)
}

func (api *deployApi) getDeployments(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Get deployments for %s/%s", workspace, name)
	deployments, _ := api.deploymentServ.ReadAll(workspace, name)
	c.JSON(http.StatusOK, deployments)
}

func (api *deployApi) createDeployment(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	zap.S().Infof("Create deployment for %s/%s", workspace, name)
	deployment, _ := api.deploymentServ.Create(workspace, name)
	c.JSON(http.StatusCreated, deployment)
}

func (api *deployApi) getDeploymentByIndex(c *gin.Context) {
	workspace := c.Param("workspace")
	name := c.Param("name")
	index := c.Param("index")
	zap.S().Infof("Get deployment %s for %s/%s", index, workspace, name)
	// Just placeholder
	c.JSON(http.StatusOK, gin.H{"index": index, "workspace": workspace, "agent": name})
}

type ActivateRequest struct {
	Agents map[string]string `json:"agents"`
}

func (api *deployApi) activateDeployment(c *gin.Context) {
	var req ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	zap.S().Infof("Activating deployment for agents: %v", req.Agents)

	var deployments []*model.Deployment
	for agentName, storePath := range req.Agents {
		deployment, err := api.deploymentServ.Create(agentName, storePath)
		if err != nil {
			zap.S().Errorf("Failed to create deployment for agent %s: %v", agentName, err)
			continue
		}
		deployments = append(deployments, deployment)
	}

	// In a real implementation, we would notify agents here.
	// For now, we return the created deployment records.
	c.JSON(http.StatusCreated, deployments)
}
