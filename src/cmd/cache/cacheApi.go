package cache

import (
	"encoding/json"
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
)

type socketApi struct {
	cache    *model.BinaryCache
	pathServ *service.StorePathSrv
	storage  objstor.ObjectStorage
}

func newApi(cache *model.BinaryCache) app.GinApi {
	var resp *socketApi
	app.Invoke(func(pathServ *service.StorePathSrv, storage objstor.ObjectStorage) {
		resp = &socketApi{cache, pathServ, storage}
	})
	return resp
}

// RegisterEndpoints implements app.GinApi.
func (s *socketApi) RegisterEndpoints(router *gin.Engine) {
	if s.cache.Access == "private" {
		zap.S().Infof("Protecting cache, access is private")
		router.Use(auth.Protect(s.cache.Token))
	}

	router.GET("/nix-cache-info", s.cacheInfo)
	router.GET("/:storeHash", s.storeHashCmd)
	router.HEAD("/:storeHash", s.storeHashCmd)

	router.PUT("/:narUuid", s.uploadData)
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
		// Your logic for .ls here
		c.JSON(200, gin.H{"type": "ls", "hash": storeHash})
		return
	}

	c.String(http.StatusNotFound, "Command not found")
}

func (s *socketApi) cacheInfo(c *gin.Context) {
	data := gin.H{"Priority": 30, "StoreDir": "/nix/store", "WantMassQuery": 1}
	resp, err := json.Marshal(data)
	if err != nil {
		c.AbortWithStatus(500)
	}

	c.Data(http.StatusOK, "application/octet-stream", resp)
}

func (s *socketApi) storeHashNarInfo(c *gin.Context, storeHash string) {
	path, err := s.pathServ.Read(storeHash, s.cache.Name)
	if err != nil {
		zap.S().Errorf("Error reading store path for hash '%s', err: %v ", storeHash, err)
		c.AbortWithStatus(500)
		return
	}

	zap.S().Infof("Found cache path, %v", path)

	// path = StorePath.get(self.server.cache.name, store_hash = m.group(1))
	// if not path:
	//     self.send_response(404)
	//     self.end_headers()
	//     return
	//
	// response = path.get_narinfo().encode('utf-8')
	//
	// self.send_response(200)
	// self.send_header("Content-Type", "text/x-nix-narinfo")
	// self.send_header("Content-Length", str(len(response)))
	// self.end_headers()
	// self.wfile.write(response)

	// Your logic for .narinfo here
	resp, err := json.Marshal(path)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.Header("Content-Length", strconv.Itoa(len(resp)))
	c.Data(http.StatusOK, "text/x-nix-narinfo", resp)
}

var writeHeaders gin.HandlerFunc = func(c *gin.Context) {
	c.Header("Content-Type", "text/x-nix-narinfo")
}

func (s *socketApi) uploadData(c *gin.Context) {
	/*
		    content_length = int(self.headers['Content-Length'])

		    filename = None
		    for f in os.listdir(self.server.cache.cache_dir):
		        if m.group(1) in f:
		            filename = f

		    if not filename:
		        self.send_response(400)
		        self.end_headers()
		        return

		    body = self.rfile.read(content_length)
		    with open(os.path.join(self.server.cache.cache_dir, filename), "wb") as file:
		        file.write(body)
		    self.send_response(201)
		    self.send_header("Content-Location", "/")
		    self.end_headers()
		else:
		    self.send_response(400)
		    self.end_headers()
	*/

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
