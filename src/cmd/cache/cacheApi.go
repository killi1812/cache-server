package cache

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/app"
	_ "github.com/killi1812/go-cache-server/docs/cache"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"github.com/killi1812/go-cache-server/util/objstor"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SocketApi struct {
	cache    *model.BinaryCache
	pathServ *service.StorePathSrv
	storage  objstor.ObjectStorage
}

func newCacheApi(
	cache *model.BinaryCache,
	pathServ *service.StorePathSrv,
	storage objstor.ObjectStorage,
) app.CreateGinApi {
	return &SocketApi{cache: cache, pathServ: pathServ, storage: storage}
}

// RegisterEndpoints implements app.GinApi.
func (s *SocketApi) NewGinApi(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.CustomWrapHandler(&ginSwagger.Config{
		InstanceName: "cache",
		URL:          "doc.json",
	}, swaggerfiles.Handler))

	if s.cache.Access == "private" {
		zap.S().Infof("Protecting cache, access is private")
		router.Use(auth.Protect(s.cache.Token))
	}

	router.GET("/nix-cache-info", s.cacheInfo)
	router.GET("/log/:deriver", s.getLog)
	router.GET("/:storeHash", s.storeHashCmd)
	router.HEAD("/:storeHash", s.storeHashCmd)

	// Nix requests /nar/<hash>.nar.<compression>
	router.GET("/nar/:filename", s.downloadNar)

	// Direct upload support
	router.PUT("/:filename", s.uploadData)
}

// downloadNar godoc
//
//	@Summary		Download NAR file
//	@Description	Download a NAR file from the cache.
//	@Tags			binary-cache
//	@Produce		octet-stream
//	@Param			filename	path		string	true	"NAR filename (e.g. hash.nar.xz)"
//	@Success		200			{file}		binary
//	@Failure		404			{object}	model.ErrorResponse
//	@Router			/nar/{filename} [get]
func (s *SocketApi) downloadNar(c *gin.Context) {
	filename := c.Param("filename")
	zap.S().Infof("Downloading NAR file: %s from cache %s", filename, s.cache.Name)

	// filename format: <fileHash>.nar.<compression>
	parts := strings.Split(filename, ".nar.")
	if len(parts) < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid nar filename"})
		return
	}
	fileHash := parts[0]

	reader, err := s.storage.ReadFile(s.cache.Name, fileHash)
	if err != nil {
		zap.S().Errorf("Failed to read NAR file %s, err: %v", fileHash, err)
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
			Error: "NAR file not found",
		})
		return
	}
	defer reader.Close()

	c.DataFromReader(http.StatusOK, -1, "application/x-nix-nar", reader, nil)
}

// getLog godoc
//
//	@Summary		Get build logs
//	@Description	Get the build logs for a particular deriver.
//	@Tags			binary-cache
//	@Produce		text/plain
//	@Param			deriver	path		string	true	"Full name of the deriver"
//	@Success		200		{string}	string	"log content"
//	@Failure		404		{object}	model.ErrorResponse
//	@Router			/log/{deriver} [get]
func (s *SocketApi) getLog(c *gin.Context) {
	deriver := c.Param("deriver")
	zap.S().Infof("Reading log for deriver: %s", deriver)

	reader, err := s.storage.ReadFile(s.cache.Name, "log/"+deriver)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{Error: "log not found"})
		return
	}
	defer reader.Close()

	c.DataFromReader(http.StatusOK, -1, "text/plain", reader, nil)
}

// storeHashCmd godoc
//
//	@Summary		Get .narinfo or .ls
//	@Description	Get metadata (.narinfo) or file listing (.ls) for a store hash.
//	@Description	- .narinfo: Text file containing store path metadata (NAR hash, size, references, signature).
//	@Description	- .ls: JSON file containing the internal file listing of the NAR.
//	@Tags			binary-cache
//	@Produce		text/plain,json
//	@Param			storeHash	path		string				true	"Store hash with extension (.narinfo or .ls)"
//	@Success		200			{string}	string				"narinfo metadata content"
//	@Success		200			{object}	map[string]any		"File listing JSON"
//	@Failure		404			{object}	model.ErrorResponse	"Metadata or listing not found"
//	@Router			/{storeHash} [get]
func (s *SocketApi) storeHashCmd(c *gin.Context) {
	filename := c.Param("storeHash")

	if before, ok := strings.CutSuffix(filename, ".narinfo"); ok {
		storeHash := before
		s.storeHashNarInfo(c, storeHash)
		return
	}

	if before, ok := strings.CutSuffix(filename, ".ls"); ok {
		storeHash := before
		zap.S().Infof("list store hash: '%s'", storeHash)

		// Attempt to read .ls file from storage
		reader, err := s.storage.ReadFile(s.cache.Name, storeHash+".ls")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{Error: "ls not found"})
			return
		}
		defer reader.Close()

		c.DataFromReader(http.StatusOK, -1, "application/json", reader, nil)
		return
	}

	c.AbortWithStatusJSON(http.StatusNotFound, model.ErrorResponse{
		Error: "Command not found",
	})
}

// cacheInfo godoc
//
//	@Summary		Get nix-cache-info
//	@Description	Get information about the nix store configuration.
//	@Tags			binary-cache
//	@Produce		text/plain
//	@Success		200	{string}	string	"StoreDir: /nix/store..."
//	@Router			/nix-cache-info [get]
func (s *SocketApi) cacheInfo(c *gin.Context) {
	resp := fmt.Sprintf("StoreDir: /nix/store\nWantMassQuery: 1\nPriority: 30\n")
	c.Data(http.StatusOK, "text/x-nix-cache-info", []byte(resp))
}

func (s *SocketApi) storeHashNarInfo(c *gin.Context, storeHash string) {
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
		c.AbortWithStatusJSON(500, model.ErrorResponse{
			Error: "failed to generate narinfo",
		})
		return
	}

	c.Header("Content-Length", strconv.Itoa(len(resp)))
	c.Data(http.StatusOK, "text/x-nix-narinfo", []byte(resp))
}

// uploadData godoc
//
//	@Summary		Upload NAR data (Direct)
//	@Description	Upload raw NAR data for a given filename (usually hash).
//	@Tags			binary-cache
//	@Accept			octet-stream
//	@Param			filename	path	string	true	"Filename or Hash"
//	@Success		201
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/{filename} [put]
func (s *SocketApi) uploadData(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.AbortWithStatusJSON(400, model.ErrorResponse{
			Error: "missing filename",
		})
		return
	}

	err := s.storage.WriteFile(s.cache.Name, filename, c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(500, model.ErrorResponse{
			Error: "failed to save to storage",
		})
		return
	}

	c.Header("Content-Location", "/")
	c.AbortWithStatus(http.StatusCreated)
}
