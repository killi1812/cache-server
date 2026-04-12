package cache

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	_ "github.com/killi1812/go-cache-server/docs"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/objstor"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	router.GET("/swagger/*any", ginSwagger.CustomWrapHandler(&ginSwagger.Config{URL: "/swagger/doc.json"}, swaggerfiles.Handler))

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

// downloadNar godoc
//
//	@Summary		Download NAR file
//	@Description	Download a NAR file from the cache.
//	@Tags			binary-cache
//	@Produce		octet-stream
//	@Param			filename	path	string	true	"NAR filename (with or without .nar extension)"
//	@Success		200			{file}	binary
//	@Failure		404
//	@Router			/nar/{filename} [get]
func (s *socketApi) downloadNar(c *gin.Context) {
	filename := c.Param("filename")
	zap.S().Infof("Downloading NAR file: %s from cache %s", filename, s.cache.Name)

	// Strip .nar extension if present
	fileHash := strings.TrimSuffix(filename, ".nar")

	reader, err := s.storage.ReadFile(s.cache.Name, fileHash)
	if err != nil {
		zap.S().Errorf("Failed to read NAR file %s, err: %v", fileHash, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer reader.Close()

	// Nix usually expects application/x-nix-archive or octet-stream
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// storeHashCmd godoc
//
//	@Summary		Get .narinfo or .ls
//	@Description	Get metadata (.narinfo) or file listing (.ls) for a store hash.
//	@Tags			binary-cache
//	@Produce		text/plain,json
//	@Param			storeHash	path		string				true	"Store hash with extension (.narinfo or .ls)"
//	@Success		200			{string}	string				"narinfo content"
//	@Success		200			{object}	map[string]string	"ls json"
//	@Failure		404
//	@Router			/{storeHash} [get]
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

// cacheInfo godoc
//
//	@Summary		Get nix-cache-info
//	@Description	Get information about the nix store configuration.
//	@Tags			binary-cache
//	@Produce		text/plain
//	@Success		200	{string}	string	"StoreDir: /nix/store..."
//	@Router			/nix-cache-info [get]
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

// uploadData godoc
//
//	@Summary		Upload NAR data
//	@Description	Upload raw NAR data for a given UUID.
//	@Tags			binary-cache
//	@Accept			octet-stream
//	@Param			narUuid	path	string	true	"NAR UUID"
//	@Success		201
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/{narUuid} [put]
func (s *socketApi) uploadData(c *gin.Context) {
	fileHash := c.Param("narUuid")
	if fileHash == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "missing file hash"})
		return
	}

	// 3. Stream the body to Object Storage
	// c.Request.Body is an io.ReadCloser. We stream it to avoid RAM spikes.
	// We no longer check for storePath in DB here because it might not exist yet during multipart upload.
	err := s.storage.WriteFile(s.cache.Name, fileHash, c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to save to storage"})
		return
	}

	c.Header("Content-Location", "/")
	c.AbortWithStatus(http.StatusCreated)
}
