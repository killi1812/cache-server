package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/killi1812/go-cache-server/api"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type ApiTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

func (suite *ApiTestSuite) SetupTest() {
	app.Test()
	gin.SetMode(gin.TestMode)

	config.Config = config.NewConfig()
	config.Config.CacheServer.Database = "file:" + suite.T().Name() + "?mode=memory&cache=shared"
	config.Config.CacheServer.CacheDir = suite.T().TempDir()

	app.Provide(db.New)
	app.Provide(objstor.New)
	app.Provide(service.NewHub)
	app.Provide(service.NewAgentSrv)
	app.Provide(service.NewCacheSrv)
	app.Provide(service.NewStorePathSrv)
	app.Provide(service.NewWorkspaceSrv)
	app.Provide(service.NewDeploymentSrv)

	// Provide APIs with names as in main.go
	app.Provide(api.NewApi, dig.Name("management"))
	app.Provide(api.NewDeployWsApi, dig.Name("deploy"))

	app.Invoke(func(database *gorm.DB) {
		db.Migration(database)
	})

	suite.router = gin.Default()
	suite.router.RedirectTrailingSlash = false

	var managementApi app.CreateGinApi
	app.Invoke(func(p struct {
		dig.In
		Api app.CreateGinApi `name:"management"`
	},
	) {
		managementApi = p.Api
	})
	managementApi.NewGinApi(suite.router)

	var err error
	suite.token, err = auth.GenerateJwt("test-user")
	assert.NoError(suite.T(), err)
}

func (suite *ApiTestSuite) requestWithToken(method, path string, body any, token string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *ApiTestSuite) request(method, path string, body any) *httptest.ResponseRecorder {
	return suite.requestWithToken(method, path, body, suite.token)
}

func (suite *ApiTestSuite) TestCacheHandlers() {
	t := suite.T()
	var cache *model.BinaryCache
	app.Invoke(func(s *service.CacheSrv) {
		var err error
		cache, err = s.Create(service.CreateCacheArgs{Name: "c-handlers", Port: 9005, Token: "t5"})
		assert.NoError(t, err)
		os.MkdirAll(filepath.Join(config.Config.CacheServer.CacheDir, "c-handlers"), 0o755)
		cache.Access = model.Public
		cache.PublicKey = "pub1"
		cache.URL = "http://localhost:9005"
		_, err = s.Update("c-handlers", *cache)
		assert.NoError(t, err)
	})

	t.Run("name - Success", func(t *testing.T) {
		w := suite.request("GET", "/api/v1/cache/c-handlers", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("name - Cache Not Found", func(t *testing.T) {
		w := suite.request("GET", "/api/v1/cache/nonexistent", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("narinfo - Success", func(t *testing.T) {
		hashes := []string{"hash-missing"}
		w := suite.request("POST", "/api/v1/cache/c-handlers/narinfo", hashes)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("createNar - Success", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar?compression=xz", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("redirect - Success", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar/uuid-1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("abortNar - Success", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar/uuid-1/abort", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func (suite *ApiTestSuite) TestDeployHandlers() {
	setup := func() {
		app.Invoke(func(cs *service.CacheSrv, ws *service.WorkspaceSrv, as *service.AgentSrv) {
			cs.Create(service.CreateCacheArgs{Name: "c-deploy-1", Port: 9001, Token: "t1"})
			ws.Create(service.WorkspaceCreateArgs{WorkspaceName: "w1", BinaryCacheName: "c-deploy-1", Token: "tw1"})
			as.Create(service.AgentCreateArgs{AgentName: "a1", WorkspaceName: "w1", Token: "ta1"})
		})
	}

	suite.T().Run("Workspace Lifecycle", func(t *testing.T) {
		setup()
		w := suite.request("GET", "/api/v1/deploy/workspace/w1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	suite.T().Run("Agent Lifecycle", func(t *testing.T) {
		setup()
		w := suite.request("GET", "/api/v1/deploy/agent/w1/a1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	suite.T().Run("Deployment Lifecycle", func(t *testing.T) {
		setup()
		w := suite.request("POST", "/api/v1/deploy/agent/w1/a-deploy", nil)
		assert.Equal(t, http.StatusCreated, w.Code)
		w = suite.request("POST", "/api/v1/deploy/deployment/w1/a-deploy", nil)
		assert.Equal(t, http.StatusCreated, w.Code)
		var dep model.Deployment
		json.Unmarshal(w.Body.Bytes(), &dep)
		w = suite.request("GET", "/api/v1/deploy/deployment/w1/a-deploy", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	suite.T().Run("Activate Deployment", func(t *testing.T) {
		setup()
		activateReq := map[string]any{"agents": map[string]string{"a1": "/nix/store/pkg"}}
		w := suite.request("POST", "/api/v1/deploy/activate", activateReq)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func (suite *ApiTestSuite) TestCacheInfo() {
	app.Invoke(func(s *service.CacheSrv) {
		s.Create(service.CreateCacheArgs{Name: "c-unique-2", Port: 9002, Token: "t2"})
	})
	w := suite.request("GET", "/api/v1/cache/c-unique-2", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ApiTestSuite) TestDeploymentActivation() {
	app.Invoke(func(cs *service.CacheSrv, ws *service.WorkspaceSrv, as *service.AgentSrv) {
		cs.Create(service.CreateCacheArgs{Name: "c-deploy-act", Port: 9010, Token: "tact"})
		ws.Create(service.WorkspaceCreateArgs{WorkspaceName: "wact", BinaryCacheName: "c-deploy-act", Token: "twact"})
		as.Create(service.AgentCreateArgs{AgentName: "aact", WorkspaceName: "wact", Token: "taact"})
	})
	activateReq := map[string]any{"agents": map[string]string{"aact": "/nix/store/hash-pkg"}}
	w := suite.request("POST", "/api/v1/deploy/activate", activateReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ApiTestSuite) TestMultipartNarCompletion() {
	app.Invoke(func(s *service.CacheSrv) {
		s.Create(service.CreateCacheArgs{Name: "c-multipart", Port: 9003, Token: "t3"})
	})
	narUuid := "00000000-0000-0000-0000-000000000001"
	completeReq := map[string]any{
		"narInfoCreate": map[string]any{
			"cStoreHash": "hash-mp", "cStoreSuffix": "suffix-mp", "cNarHash": "narhash-mp",
			"cNarSize": 1234, "cFileHash": "filehash-mp", "cFileSize": 5678,
			"cReferences": []string{"ref1"}, "cDeriver": "deriver-mp", "cSig": "sig-mp",
		},
	}
	w := suite.request("POST", "/api/v1/cache/c-multipart/multipart-nar/"+narUuid+"/complete", completeReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

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

	// Provide APIs with names
	app.Provide(api.NewApi, dig.Name("management"))
	app.Provide(api.NewDeployWsApi, dig.Name("deploy"))

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
	var wsApi app.CreateGinApi
	app.Invoke(func(p struct {
		dig.In
		Api app.CreateGinApi `name:"deploy"`
	},
	) {
		wsApi = p.Api
	})
	wsApi.NewGinApi(router)

	s := httptest.NewServer(router)
	defer s.Close()

	t.Run("Agent Connection and Registration", func(t *testing.T) {
		u := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws?name=test-agent&token=a-token"
		client, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer client.Close()

		var msg map[string]any
		err = client.ReadJSON(&msg)
		assert.NoError(t, err)
		assert.Equal(t, "AgentRegistered", msg["method"])
	})

	t.Run("Deployment Feedback", func(t *testing.T) {
		var dep *model.Deployment
		app.Invoke(func(s *service.DeploymentSrv) {
			dep, _ = s.Create("test-agent", "/nix/store/abc")
		})

		u := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws-deployment"
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
