package api

import (
	"fmt"
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
func (api *deployWsApi) wsHandler(c *gin.Context) {
	name := c.GetHeader("name")
	token := c.GetHeader("Authorization")
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	zap.S().Debugf("Name: %s, Token: %s", name, token)

	if name == "" || token == "" {
		name = c.Query("name")
		token = c.Query("token")
		zap.S().Debugf("Name: %s, Token: %s", name, token)
	}

	if name == "" || token == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "missing agent name or token"})
		return
	}

	agent, err := api.agentServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read agent '%s': %v", name, err)
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	if agent.Token != token {
		zap.S().Errorf("Token mismatch for agent '%s'", name)
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
		conn.WriteJSON(model.ErrorResponse{Error: "misconfigured agent relations"})
		conn.Close()
		return
	}

	// Send AgentRegistered message
	cacheURI := agent.Workspace.BinaryCache.URL
	if strings.Contains(cacheURI, "localhost") {
		domain := strings.Split(c.Request.Host, ":")[0]
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		cacheURI = fmt.Sprintf("%s://%s.%s", scheme, agent.Workspace.BinaryCache.Name, domain)
	}

	// Strip prefix from public key for WebSocket (agent adds its own)
	pubKeyParts := strings.Split(agent.Workspace.BinaryCache.PublicKey, ":")
	rawPubKey := pubKeyParts[len(pubKeyParts)-1]
	keyName := agent.Workspace.BinaryCache.Name + ".localhost-1"
	fullKey := keyName + ":" + rawPubKey

	regMsg := map[string]any{
		"agent": agent.Uuid.String(),
		"command": map[string]any{
			"contents": map[string]any{
				"id":             agent.Uuid.String(),
				"agentId":        agent.Uuid.String(),
				"agentName":      name,
				"agentVersion":   "1.9.1", // Match agent version
				"agentPublicKey": "",      // Placeholder
				"cache": map[string]any{
					"name":              agent.Workspace.BinaryCache.Name,
					"cacheName":         agent.Workspace.BinaryCache.Name,
					"uri":               cacheURI,
					"cacheUri":          cacheURI,
					"publicKey":         rawPubKey,
					"publicSigningKeys": []string{fullKey},
					"isPublic":          agent.Workspace.BinaryCache.Access == "public",
					"githubUsername":    "",
				},
			},
			"tag": "AgentRegistered",
		},
		"id":     "00000000-0000-0000-0000-000000000000",
		"method": "AgentRegistered",
	}
	zap.S().Infof("Sending AgentRegistered to agent '%s': %+v", name, regMsg)
	conn.WriteJSON(regMsg)

	defer func() {
		zap.S().Infof("Closing WebSocket connection for agent '%s'", name)
		api.hub.Unregister(name)
		conn.Close()
	}()

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			zap.S().Debugf("WebSocket read error for agent '%s': %v", name, err)
			break
		}
		zap.S().Infof("Received WebSocket message from agent '%s': %+v", name, msg)
		method, _ := msg["method"].(string)
		if method == "DeploymentFinished" {
			api.processDeploymentFinished(msg)
		}
	}
}

// deploymentHandler godoc
//
//	@Summary		Deployment status WebSocket
//	@Description	Channel for agents to report deployment completion.
//	@Tags			deployment
//	@Success		101
//	@Router			/deploy/ws-deployment [get]
func (api *deployWsApi) deploymentHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.S().Errorf("Failed to upgrade deployment websocket: %v", err)
		return
	}
	defer conn.Close()

	zap.S().Info("New deployment status WebSocket connection")

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			zap.S().Debugf("Deployment status WebSocket read error: %v", err)
			break
		}
		zap.S().Infof("Received deployment status message: %+v", msg)
		method, _ := msg["method"].(string)
		if method == "DeploymentFinished" {
			api.processDeploymentFinished(msg)
			conn.Close() // Match Python await websocket.close()
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
func (api *deployWsApi) logHandler(c *gin.Context) {
	id := c.Param("id")
	zap.S().Infof("New log streaming connection for deployment %s", id)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.S().Errorf("Failed to upgrade log websocket for %s: %v", id, err)
		return
	}
	defer conn.Close()

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			zap.S().Debugf("Log WebSocket read error for %s: %v", id, err)
			break
		}
		zap.S().Infof("Raw log message for %s: %+v", id, msg)

		if line, ok := msg["line"].(string); ok {
			zap.S().Infof("Agent Log [%s]: %s", id, line)
			if line == "Successfully activated the deployment." || strings.Contains(line, "Failed to activate the deployment.") {
				break
			}
		}
	}
	zap.S().Infof("Log streaming finished for deployment %s", id)
}

func (api *deployWsApi) processDeploymentFinished(msg map[string]any) {
	zap.S().Debugf("Incoming Message: %+v", msg)
	command, _ := msg["command"].(map[string]any)
	id, _ := command["id"].(string)
	success, _ := command["hasSucceeded"].(bool)

	zap.S().Infof("Agent reported deployment %s finished (Success: %v)", id, success)

	status := model.DeploymentSuccess
	if !success {
		status = model.DeploymentFailed
	}

	err := api.deploymentServ.UpdateStatus(id, status)
	if err != nil {
		zap.S().Errorf("Failed to update deployment %s status: %v", id, err)
	} else {
		zap.S().Infof("Database updated: Deployment %s is now %s", id, status)
	}
}
