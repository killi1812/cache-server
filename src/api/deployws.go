package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Production should restrict this
	},
}

// wsHandler godoc
//
//	@Summary		Agent WebSocket registration
//	@Description	Long-lived WebSocket connection for agents to receive deployment commands.
//	@Tags			deployment
//	@Param			name	query	string	true	"Agent Name"
//	@Param			token	query	string	true	"Agent Token"
//	@Success		101
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		401	{object}	model.ErrorResponse
//	@Router			/deploy/ws [get]
func (api *deployApi) wsHandler(c *gin.Context) {
	name := c.GetHeader("name")
	token := c.GetHeader("Authorization")
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	if name == "" || token == "" {
		name = c.Query("name")
		token = c.Query("token")
	}

	if name == "" || token == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "missing agent name or token"})
		return
	}

	agent, err := api.agentServ.Read(name)
	if err != nil || agent.Token != token {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.S().Errorf("Failed to upgrade websocket: %v", err)
		return
	}

	api.hub.Register(name, conn)

	// Verify relations exist
	if agent.Workspace == nil || agent.Workspace.BinaryCache == nil {
		zap.S().Errorf("Agent '%s' missing workspace or binary cache relations", name)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "misconfigured agent relations"})
		return
	}

	// Send AgentRegistered message
	regMsg := map[string]any{
		"agent": agent.Uuid.String(),
		"command": map[string]any{
			"contents": map[string]any{
				"cache": map[string]any{
					"name": agent.Workspace.BinaryCache.Name,
					"uri":  agent.Workspace.BinaryCache.URL,
				},
				"id": agent.Uuid.String(),
			},
			"tag": "AgentRegistered",
		},
		"id":     "00000000-0000-0000-0000-000000000000", // TODO: what to do with this id
		"method": "AgentRegistered",
	}
	conn.WriteJSON(regMsg)

	// Keep connection alive
	go func() {
		defer func() {
			api.hub.Unregister(name)
			conn.Close()
		}()
		for {
			var msg map[string]any
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}
			method, _ := msg["method"].(string)
			if method == "DeploymentFinished" {
				api.processDeploymentFinished(msg)
			}
		}
	}()
}

// deploymentHandler godoc
//
//	@Summary		Deployment status WebSocket
//	@Description	Channel for agents to report deployment completion.
//	@Tags			deployment
//	@Success		101
//	@Router			/deploy/ws-deployment [get]
func (api *deployApi) deploymentHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		method, _ := msg["method"].(string)
		if method == "DeploymentFinished" {
			api.processDeploymentFinished(msg)
			break
		}
	}
}

// logHandler godoc
//
//	@Summary		Deployment log WebSocket
//	@Description	Stream real-time logs from the agent during deployment.
//	@Tags			deployment
//	@Success		101
//	@Router			/deploy/log/ [get]
func (api *deployApi) logHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		line, _ := msg["line"].(string)
		zap.S().Infof("Agent Log: %s", line)
		if line == "Successfully activated the deployment." || strings.Contains(line, "Failed to activate the deployment.") {
			break
		}
	}
}

func (api *deployApi) processDeploymentFinished(msg map[string]any) {
	command, _ := msg["command"].(map[string]any)
	id, _ := command["id"].(string)
	success, _ := command["hasSucceeded"].(bool)

	status := model.DeploymentSuccess
	if !success {
		status = model.DeploymentFailed
	}

	_ = api.deploymentServ.UpdateStatus(id, status)
}
