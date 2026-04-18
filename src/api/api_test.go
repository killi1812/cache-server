package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
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
	config.Config.CacheServer.CacheDir = suite.T().TempDir()

	// Provide dependencies
	app.Provide(db.New)
	app.Provide(objstor.New)
	app.Provide(service.NewHub)
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
	suite.router.RedirectTrailingSlash = false
	managementApi := api.NewApi()
	managementApi.NewGinApi(suite.router)

	// Generate a valid token for tests
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

func (suite *ApiTestSuite) TestDeployHandlers() {
	t := suite.T()

	// 1. Setup Common Prerequisite: Cache -> Workspace -> Agent
	app.Invoke(func(cs *service.CacheSrv, ws *service.WorkspaceSrv, as *service.AgentSrv) {
		cs.Create(service.CreateCacheArgs{Name: "c-deploy-1", Port: 9001, Token: "t1"})
		ws.Create(service.WorkspaceCreateArgs{WorkspaceName: "w1", BinaryCacheName: "c-deploy-1", Token: "tw1"})
		as.Create(service.AgentCreateArgs{AgentName: "a1", WorkspaceName: "w1", Token: "ta1"})
	})

	t.Run("Workspace Lifecycle", func(t *testing.T) {
		// Create Workspace - Invalid Body
		w := suite.request("POST", "/api/v1/deploy/workspace", "invalid")
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Get Workspace
		w = suite.request("GET", "/api/v1/deploy/workspace/w1", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		// Get Workspace - Not Found
		w = suite.request("GET", "/api/v1/deploy/workspace/nonexistent", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Agent Lifecycle", func(t *testing.T) {
		// Get Agent
		w := suite.request("GET", "/api/v1/deploy/agent/w1/a1", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		// Get Agent - Not Found in DB
		w = suite.request("GET", "/api/v1/deploy/agent/w1/nonexistent", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Get Agent - Wrong Workspace
		app.Invoke(func(s *service.WorkspaceSrv) {
			s.Create(service.WorkspaceCreateArgs{WorkspaceName: "w2", BinaryCacheName: "c-deploy-1", Token: "t2"})
		})
		w = suite.request("GET", "/api/v1/deploy/agent/w2/a1", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		// List Agents
		w = suite.request("GET", "/api/v1/deploy/workspace/w1/agents", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var agents []any
		json.Unmarshal(w.Body.Bytes(), &agents)
		assert.NotEmpty(t, agents)
	})

	t.Run("Deployment Lifecycle", func(t *testing.T) {
		// Create a fresh agent for this test
		w := suite.request("POST", "/api/v1/deploy/agent/w1/a-deploy", nil)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Create Deployment
		w = suite.request("POST", "/api/v1/deploy/deployment/w1/a-deploy", nil)
		assert.Equal(t, http.StatusCreated, w.Code)
		var dep model.Deployment
		json.Unmarshal(w.Body.Bytes(), &dep)

		// Get Deployment by ID - Service is placeholder returning nil, so expect 404
		w = suite.request("GET", "/api/v1/deploy/deployment/"+dep.Uuid.String(), nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Get Deployments for Agent
		w = suite.request("GET", "/api/v1/deploy/deployment/w1/a-deploy", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		// Get Deployment by Index
		w = suite.request("GET", "/api/v1/deploy/deployment/w1/a-deploy/0", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		t.Run("Deployment - Missing Params", func(t *testing.T) {
			// NOTE:
			// Gin router usually doesn't match if params are missing entirely for these routes,
			// but we can test if we hit them somehow or if they are empty strings.
			// Actually, with RedirectTrailingSlash=false, some might match.

			// getDeployments missing name
			w := suite.request("GET", "/api/v1/deploy/deployment/w1/", nil)
			// NOTE:
			// Depending on gin, this might 404 if it doesn't match /:workspace/:name
			// If it matches /:workspace, it's getDeployment
			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		})
	})

	t.Run("Activate Deployment", func(t *testing.T) {
		activateReq := map[string]any{
			"agents": map[string]string{
				"a1": "/nix/store/hash-pkg",
			},
		}
		w := suite.request("POST", "/api/v1/deploy/activate", activateReq)
		assert.Equal(t, http.StatusOK, w.Code)

		// Invalid Body
		w = suite.request("POST", "/api/v1/deploy/activate", "invalid")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Delete Agent
		w := suite.request("DELETE", "/api/v1/deploy/agent/w1/a1", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Delete Agent - Not Found
		w = suite.request("DELETE", "/api/v1/deploy/agent/w1/a1", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Delete Workspace
		w = suite.request("DELETE", "/api/v1/deploy/workspace/w1", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func (suite *ApiTestSuite) TestCacheHandlers() {
	t := suite.T()

	// 1. Create Cache
	var cache *model.BinaryCache
	app.Invoke(func(s *service.CacheSrv) {
		var err error
		cache, err = s.Create(service.CreateCacheArgs{Name: "c-handlers", Port: 9005, Token: "t5"})
		assert.NoError(t, err)

		// Ensure storage directory exists
		os.MkdirAll(filepath.Join(config.Config.CacheServer.CacheDir, "c-handlers"), 0o755)

		// Set public and other fields via Update if needed (though Create sets some)
		cache.Access = model.Public
		cache.PublicKey = "pub1"
		cache.URL = "http://localhost:9005"
		_, err = s.Update("c-handlers", *cache)
		assert.NoError(t, err)
	})

	t.Run("name - Success", func(t *testing.T) {
		w := suite.request("GET", "/api/v1/cache/c-handlers", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var nameResp map[string]any
		json.Unmarshal(w.Body.Bytes(), &nameResp)
		assert.Equal(t, "c-handlers", nameResp["name"])
	})

	t.Run("name - Cache Not Found", func(t *testing.T) {
		w := suite.request("GET", "/api/v1/cache/nonexistent", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code) // Current implementation returns 500 on read error
	})

	t.Run("name - Private Cache Access", func(t *testing.T) {
		// Create a private cache
		token, _ := auth.GenerateJwt("test-user")
		app.Invoke(func(s *service.CacheSrv) {
			privCache, _ := s.Create(service.CreateCacheArgs{Name: "c-private", Port: 9006, Token: token})
			privCache.Access = model.Private
			s.Update("c-private", *privCache)
		})

		// Request with NO token (should fail 401)
		req, _ := http.NewRequest("GET", "/api/v1/cache/c-private", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Request with VALID token (should pass)
		w = suite.requestWithToken("GET", "/api/v1/cache/c-private", nil, token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("narinfo - Success", func(t *testing.T) {
		hashes := []string{"hash-missing", "hash-exists"}
		w := suite.request("POST", "/api/v1/cache/c-handlers/narinfo", hashes)
		assert.Equal(t, http.StatusOK, w.Code)
		var missing []string
		json.Unmarshal(w.Body.Bytes(), &missing)
		assert.Contains(t, missing, "hash-missing")
	})

	t.Run("narinfo - Invalid JSON", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/narinfo", "not-a-json-array")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("createNar - Success", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar?compression=xz", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var createNarResp map[string]string
		json.Unmarshal(w.Body.Bytes(), &createNarResp)
		assert.Contains(t, createNarResp, "narId")
	})

	t.Run("createNar - Invalid Compression", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar?compression=zip", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("redirect - Success", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar/some-uuid", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var redirectResp map[string]string
		json.Unmarshal(w.Body.Bytes(), &redirectResp)
		assert.Contains(t, redirectResp["uploadUrl"], "some-uuid")
	})

	t.Run("redirect - Cache Not Found", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/nonexistent/multipart-nar/some-uuid", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("completeNar - Success", func(t *testing.T) {
		narUuid := "complete-me"
		completeReq := map[string]any{
			"narInfoCreate": map[string]any{
				"cStoreHash":   "final-hash",
				"cStoreSuffix": "final-suffix",
			},
		}
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar/"+narUuid+"/complete", completeReq)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("completeNar - Invalid Body", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar/uuid/complete", "invalid")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("abortNar - Success", func(t *testing.T) {
		w := suite.request("POST", "/api/v1/cache/c-handlers/multipart-nar/abort-me/abort", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
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
