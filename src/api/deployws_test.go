package api

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestDeployWebSocket(t *testing.T) {
	app.Test()
	gin.SetMode(gin.TestMode)

	config.Config = config.NewConfig()
	config.Config.CacheServer.Database = "file:" + t.Name() + "?mode=memory&cache=shared"

	app.Provide(db.New)
	app.Provide(service.NewHub)
	app.Provide(service.NewAgentSrv)
	app.Provide(service.NewWorkspaceSrv)
	app.Provide(service.NewCacheSrv)
	app.Provide(service.NewDeploymentSrv)

	var database *gorm.DB
	app.Invoke(func(d *gorm.DB) {
		database = d
		db.Migration(d)
	})

	// Setup data
	cache := &model.BinaryCache{Name: "test-cache", Token: "c-token"}
	database.Create(cache)
	ws := &model.Workspace{Name: "test-ws", BinaryCacheId: cache.ID}
	database.Create(ws)
	agent := &model.Agent{Name: "test-agent", Token: "a-token", WorkspaceId: ws.ID}
	database.Create(agent)

	router := gin.Default()
	// Use RegisterEndpoints as defined in deploy.go
	deployApiInst := newDeployApi().(*deployApi)
	deployApiInst.RegisterEndpoints(router.Group("/api/v1"))

	s := httptest.NewServer(router)
	defer s.Close()

	t.Run("Agent Connection and Registration", func(t *testing.T) {
		u := "ws" + strings.TrimPrefix(s.URL, "http") + "/api/v1/deploy/ws?name=test-agent&token=a-token"
		client, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer client.Close()

		var msg map[string]any
		err = client.ReadJSON(&msg)
		assert.NoError(t, err)
		assert.Equal(t, "AgentRegistered", msg["method"])
	})

	t.Run("Deployment Feedback", func(t *testing.T) {
		dep, _ := service.NewDeploymentSrv().Create("test-agent", "/nix/store/abc")
		
		u := "ws" + strings.TrimPrefix(s.URL, "http") + "/api/v1/deploy/ws-deployment"
		client, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer client.Close()

		feedback := map[string]any{
			"method": "DeploymentFinished",
			"command": map[string]any{
				"id":           dep.Uuid.String(),
				"hasSucceeded": true,
			},
		}
		err = client.WriteJSON(feedback)
		assert.NoError(t, err)

		// Wait for processing
		var updated model.Deployment
		for i := 0; i < 20; i++ {
			database.Where("uuid = ?", dep.Uuid).First(&updated)
			if updated.Status == model.DeploymentSuccess {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		assert.Equal(t, model.DeploymentSuccess, updated.Status)
	})
}
