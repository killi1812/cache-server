package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/api"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ApiTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

func (suite *ApiTestSuite) SetupTest() {
	app.Test() // Initialize dig container
	gin.SetMode(gin.TestMode)

	// Setup Config
	config.Config = config.NewConfig()
	// Use a unique memory database name for each test to ensure absolute isolation
	suite.T().Logf("Setting up test: %s", suite.T().Name())
	config.Config.CacheServer.Database = "file:" + suite.T().Name() + "?mode=memory&cache=shared"

	// Provide dependencies
	app.Provide(db.New)
	app.Provide(objstor.New)
	app.Provide(service.NewAgentSrv)
	app.Provide(service.NewCacheSrv)
	app.Provide(service.NewStorePathSrv)
	app.Provide(service.NewWorkspaceSrv)
	app.Provide(service.NewDeploymentSrv)

	// Migrate
	app.Invoke(func(database *gorm.DB) {
		db.Migration(database)
	})

	suite.router = gin.Default()
	managementApi := api.NewApi()
	managementApi.NewGinApi(suite.router)

	// Generate a valid token for tests
	var err error
	suite.token, err = auth.GenerateJwt("test-user")
	assert.NoError(suite.T(), err)
}

func (suite *ApiTestSuite) request(method, path string, body any) *httptest.ResponseRecorder {
	var bodyReader *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *ApiTestSuite) TestWorkspaceAndAgentLifecycle() {
	t := suite.T()

	// 1. Create a Cache first (required for workspace)
	app.Invoke(func(s *service.CacheSrv) {
		_, err := s.Create(service.CreateCacheArgs{Name: "c-unique-1", Port: 9001, Token: "t1"})
		assert.NoError(t, err)
	})

	// 2. Create Workspace
	wsReq := map[string]string{"name": "w-unique-1", "cacheName": "c-unique-1"}
	w := suite.request("POST", "/api/v1/deploy/workspace", wsReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 3. Get Workspace
	w = suite.request("GET", "/api/v1/deploy/workspace/w-unique-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Create Agent
	w = suite.request("POST", "/api/v1/deploy/agent/w-unique-1/a-unique-1", nil)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 5. Get Agent
	w = suite.request("GET", "/api/v1/deploy/agent/w-unique-1/a-unique-1", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. List Agents
	w = suite.request("GET", "/api/v1/deploy/workspace/w-unique-1/agents", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var agents []any
	json.Unmarshal(w.Body.Bytes(), &agents)
	assert.Len(t, agents, 1)

	// 7. Delete Agent
	w = suite.request("DELETE", "/api/v1/deploy/agent/w-unique-1/a-unique-1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// 8. Delete Workspace
	w = suite.request("DELETE", "/api/v1/deploy/workspace/w-unique-1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func (suite *ApiTestSuite) TestMultipartNarCompletion() {
	t := suite.T()

	// 1. Create Cache
	app.Invoke(func(s *service.CacheSrv) {
		s.Create(service.CreateCacheArgs{Name: "c-multipart", Port: 9003, Token: "t3"})
	})

	// 2. Complete Multipart
	narUuid := "00000000-0000-0000-0000-000000000001"
	completeReq := map[string]any{
		"narInfoCreate": map[string]any{
			"cStoreHash":   "hash-mp",
			"cStoreSuffix": "suffix-mp",
			"cNarHash":     "narhash-mp",
			"cNarSize":     1234,
			"cFileHash":    "filehash-mp",
			"cFileSize":    5678,
			"cReferences":  []string{"ref1", "ref2"},
			"cDeriver":     "deriver-mp",
			"cSig":         "sig-mp",
		},
	}

	w := suite.request("POST", "/api/v1/cache/c-multipart/multipart-nar/"+narUuid+"/complete?uploadId=up1", completeReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Verify in DB via StorePathSrv
	app.Invoke(func(s *service.StorePathSrv) {
		path, err := s.Read("hash-mp", "c-multipart")
		assert.NoError(t, err)
		assert.NotNil(t, path)
		assert.Equal(t, "suffix-mp", path.StoreSuffix)
		assert.Equal(t, narUuid, path.FileHash) // Should match our mapping
	})
}

func (suite *ApiTestSuite) TestDeploymentActivation() {
	t := suite.T()

	// 1. Setup prerequisite: Cache -> Workspace -> Agent
	app.Invoke(func(cs *service.CacheSrv, ws *service.WorkspaceSrv, as *service.AgentSrv) {
		cs.Create(service.CreateCacheArgs{Name: "c-deploy", Port: 9004, Token: "t4"})
		ws.Create(service.WorkspaceCreateArgs{WorkspaceName: "w-deploy", BinaryCacheName: "c-deploy", Token: "tw"})
		as.Create(service.AgentCreateArgs{AgentName: "a-deploy", WorkspaceName: "w-deploy", Token: "ta"})
	})

	// 2. Activate Deployment
	activateReq := map[string]any{
		"agents": map[string]string{
			"a-deploy": "/nix/store/hash-deploy-pkg",
		},
	}

	w := suite.request("POST", "/api/v1/deploy/activate", activateReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "/nix/store/hash-deploy-pkg", resp[0]["storePath"])
	assert.Equal(t, "pending", resp[0]["status"])
}

func (suite *ApiTestSuite) TestCacheInfo() {
	t := suite.T()

	// Create Cache
	app.Invoke(func(s *service.CacheSrv) {
		s.Create(service.CreateCacheArgs{Name: "c-unique-2", Port: 9002, Token: "t2"})
	})

	w := suite.request("GET", "/api/v1/cache/c-unique-2", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "c-unique-2", resp["name"])
	assert.Contains(t, resp, "publicSigningKeys")
	assert.Equal(t, "XZ", resp["preferredCompressionMethod"])
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}
