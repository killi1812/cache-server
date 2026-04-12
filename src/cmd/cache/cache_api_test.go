package cache

import (
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
	app.Provide(service.NewStorePathSrv)
	app.Provide(service.NewCacheSrv)

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
	socketApi := newCacheApi(cache)
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
}
