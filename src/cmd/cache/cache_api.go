package cache

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/objstor"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type socketApi struct {
	cache    *model.BinaryCache
	pathServ *service.StorePathSrv
	storage  objstor.ObjectStorage
}

func newCacheApi(cache *model.BinaryCache) app.CreateGinApi {
	var resp *socketApi
	app.Invoke(func(pathServ *service.StorePathSrv, storage objstor.ObjectStorage) {
		resp = &socketApi{cache, pathServ, storage}
	})
	return resp
}

// RegisterEndpoints implements app.GinApi.
func (s *socketApi) NewGinApi(router *gin.Engine) {
	if s.cache.Access == "private" {
		zap.S().Infof("Protecting cache, access is private")
		router.Use(auth.Protect(s.cache.Token))
	}

	router.GET("/nix-cache-info", s.cacheInfo)
	router.GET("/:storeHash", s.storeHashCmd)
	// TODO: see what to do with this
	router.HEAD("/:storeHash", s.storeHashCmd)

	router.GET("/nar/:filename", s.downloadNar)

	router.PUT("/:narUuid", s.uploadData)
}

func (s *socketApi) downloadNar(c *gin.Context) {
	filename := c.Param("filename")
	zap.S().Infof("Downloading NAR file: %s", filename)

	// Strip .nar extension if present
	fileHash := strings.TrimSuffix(filename, ".nar")

	reader, err := s.storage.ReadFile(fileHash)
	if err != nil {
		zap.S().Errorf("Failed to read NAR file %s, err: %v", fileHash, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer reader.Close()

	// Nix usually expects application/x-nix-archive or octet-stream
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (s *socketApi) storeHashCmd(c *gin.Context) {
	filename := c.Param("storeHash")

	if before, ok := strings.CutSuffix(filename, ".narinfo"); ok {
		storeHash := before
		s.storeHashNarInfo(c, storeHash)
		return
	}

	if before, ok := strings.CutSuffix(filename, ".ls"); ok {
		storeHash := before
		zap.S().Infof("list store hash: '%s'", storeHash)
		// TODO: implement ls logic
		c.JSON(200, gin.H{"type": "ls", "hash": storeHash})
		return
	}

	c.String(http.StatusNotFound, "Command not found")
}

func (s *socketApi) cacheInfo(c *gin.Context) {
	resp := fmt.Sprintf("StoreDir: /nix/store\nWantMassQuery: 1\nPriority: 30\n")
	c.Data(http.StatusOK, "text/x-nix-cache-info", []byte(resp))
}

func (s *socketApi) storeHashNarInfo(c *gin.Context, storeHash string) {
	path, err := s.pathServ.Read(storeHash, s.cache.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.S().Errorf("Store path not found hash '%s', err: %v ", storeHash, err)
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			zap.S().Errorf("Error reading store path for hash '%s', err: %v ", storeHash, err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	zap.S().Infof("Found cache path, %v", path)

	resp, err := s.pathServ.GenerateNarInfo(path, s.cache.SecretKey)
	if err != nil {
		zap.S().Errorf("Failed to generate narinfo: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.Header("Content-Length", strconv.Itoa(len(resp)))
	c.Data(http.StatusOK, "text/x-nix-narinfo", []byte(resp))
}

func (s *socketApi) uploadData(c *gin.Context) {
	fileHash := c.Param("narUuid")
	if fileHash == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "missing file hash"})
		return
	}

	storePath, err := s.pathServ.Read(fileHash, s.cache.Name)
	if err != nil || storePath == nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "file record not found in database"})
		return
	}

	// 3. Stream the body to Object Storage
	// c.Request.Body is an io.ReadCloser. We stream it to avoid RAM spikes.
	err = s.storage.WriteFile(storePath.FileHash, c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to save to storage"})
		return
	}

	c.Header("Content-Location", "/")
	c.AbortWithStatus(http.StatusCreated)
}
