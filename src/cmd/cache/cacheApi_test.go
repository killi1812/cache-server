package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSocketApi(t *testing.T) {
	app.Test()
	gin.SetMode(gin.TestMode)

	config.Config = config.NewConfig()
	config.Config.CacheServer.Database = "file:" + t.Name() + "?mode=memory&cache=shared"
	config.Config.CacheServer.CacheDir = t.TempDir()

	app.Provide(db.New)
	app.Provide(objstor.New)
	app.Provide(service.NewHub)
	app.Provide(service.NewAgentSrv)
	app.Provide(service.NewStorePathSrv)
	app.Provide(service.NewCacheSrv)
	app.Provide(service.NewDeploymentSrv)

	var database *gorm.DB
	app.Invoke(func(d *gorm.DB) {
		database = d
		db.Migration(d)
	})

	cache := &model.BinaryCache{
		Name:      "test-socket",
		Port:      9999,
		SecretKey: "test-socket:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}
	database.Create(cache)

	// Create a store path for testing .narinfo
	sp := &model.StorePath{
		StoreHash:     "hash1",
		StoreSuffix:   "suffix1",
		BinaryCacheId: cache.ID,
	}
	database.Create(sp)

	router := gin.Default()
	var socketApi app.CreateGinApi
	app.Invoke(func(
		pathServ *service.StorePathSrv,
		agentServ *service.AgentSrv,
		deploymentServ *service.DeploymentSrv,
		storage objstor.ObjectStorage,
		hub *service.Hub,
	) {
		socketApi = newCacheApi(cache, pathServ, agentServ, deploymentServ, storage, hub)
	})
	socketApi.NewGinApi(router)

	t.Run("Get nix-cache-info", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nix-cache-info", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "StoreDir: /nix/store")
		assert.Equal(t, "text/x-nix-cache-info", w.Header().Get("Content-Type"))
	})

	t.Run("Get .narinfo", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hash1.narinfo", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "StorePath: /nix/store/hash1-suffix1")
		assert.Equal(t, "text/x-nix-narinfo", w.Header().Get("Content-Type"))
	})

	t.Run("Download NAR", func(t *testing.T) {
		// 1. Create a dummy file in the "storage"
		storageDir := config.Config.CacheServer.CacheDir
		fileName := "dummyhash"
		content := []byte("dummy nar content")
		
		// Ensure directory exists
		os.MkdirAll(filepath.Join(storageDir, cache.Name), 0755)
		os.WriteFile(filepath.Join(storageDir, cache.Name, fileName), content, 0644)
		defer os.Remove(filepath.Join(storageDir, cache.Name, fileName))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nar/"+fileName+".nar", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, content, w.Body.Bytes())
		assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	})

	t.Run("HEAD .narinfo", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/hash1.narinfo", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get .ls", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hash1.ls", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "\"type\":\"ls\"")
	})

	t.Run("Get non-existent .narinfo", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nonexistent.narinfo", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Download non-existent NAR", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nar/nonexistent.nar", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Upload NAR", func(t *testing.T) {
		content := []byte("uploaded content")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/uploadhash", bytes.NewBuffer(content))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Verify file exists in storage
		storagePath := filepath.Join(config.Config.CacheServer.CacheDir, cache.Name, "uploadhash")
		_, err := os.Stat(storagePath)
		assert.NoError(t, err)
		
		savedContent, _ := os.ReadFile(storagePath)
		assert.Equal(t, content, savedContent)
	})

	t.Run("Private Cache Unauthorized", func(t *testing.T) {
		token, _ := auth.GenerateJwt("test-user")
		privateCache := &model.BinaryCache{
			Name:   "private-cache",
			Access: "private",
			Token:  token,
		}
		database.Create(privateCache)

		privateRouter := gin.Default()
		var privateApi app.CreateGinApi
		app.Invoke(func(
			pathServ *service.StorePathSrv,
			agentServ *service.AgentSrv,
			deploymentServ *service.DeploymentSrv,
			storage objstor.ObjectStorage,
			hub *service.Hub,
		) {
			privateApi = newCacheApi(privateCache, pathServ, agentServ, deploymentServ, storage, hub)
		})
		privateApi.NewGinApi(privateRouter)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nix-cache-info", nil)
		privateRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
