package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/auth"
	"go.uber.org/zap"
)

type CacheCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Port      int    `json:"port" binding:"required"`
	Retention int    `json:"retention"`
	Token     string `json:"token"`
}

// create godoc
//
//	@Summary		Create binary cache
//	@Description	Create a new binary cache.
//	@Tags			cache
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CacheCreateRequest	true	"Cache details"
//	@Success		201		{object}	model.BinaryCache
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/cache [post]
func (api *cacheApi) create(c *gin.Context) {
	var req CacheCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	zap.S().Infof("Trying to create cache '%s' on port %d", req.Name, req.Port)

	token := req.Token
	if token == "" {
		var err error
		token, err = auth.GenerateJwt(req.Name)
		if err != nil {
			zap.S().Errorf("Failed to generate token, err: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "failed to generate token",
			})
			return
		}
	}

	args := service.CreateCacheArgs{
		Name:      req.Name,
		Port:      req.Port,
		Retention: req.Retention,
		Token:     token,
	}

	cache, err := api.cacheServ.Create(args)
	if err != nil {
		zap.S().Errorf("Failed to create cache: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to create cache",
		})
		return
	}

	// Create storage directory
	_, err = api.storage.CreateDir(req.Name)
	if err != nil {
		zap.S().Errorf("Failed to create storage for cache '%s': %v", req.Name, err)
		// TODO: rollback DB create
	}

	c.JSON(http.StatusCreated, cache)
}

// name godoc
//
//	@Summary		Get cache info
//	@Description	Get detailed information about a binary cache.
//	@Tags			cache
//	@Produce		json
//	@Param			name	path		string	true	"Cache Name"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/cache/{name} [get]
func (api *cacheApi) name(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to read cache '%s'", name)

	cache, err := api.cacheServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache '%s', err: %v", name, err)
		c.AbortWithStatusJSON(500, model.ErrorResponse{
			Error: "failed to read cache",
		})
		return
	}

	if cache.Access == model.Private {
		// Not exactly middleware but manual check for now
		auth.Protect(cache.Token)(c)
		if c.IsAborted() {
			return
		}
	}

	// Cachix-compliant response
	response := gin.H{
		"githubUsername":             "",
		"isPublic":                   cache.Access == model.Public,
		"name":                       cache.Name,
		"permission":                 "Admin", // Default for now
		"preferredCompressionMethod": "XZ",
		"publicSigningKeys":          []string{cache.PublicKey},
		"uri":                        cache.URL,
	}

	c.JSON(http.StatusOK, response)
}

// narinfo godoc
//
//	@Summary		Get missing narinfo hashes
//	@Description	Returns a list of hashes from the input that are missing in the cache.
//	@Tags			cache
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string		true	"Cache Name"
//	@Param			hashes	body		[]string	true	"Hashes to check"
//	@Success		200		{array}		string
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/cache/{name}/narinfo [post]
func (api *cacheApi) narinfo(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to retrive missing narinfo '%s'", name)

	var incomingHashes []string

	if err := c.ShouldBindJSON(&incomingHashes); err != nil {
		c.AbortWithStatusJSON(400, model.ErrorResponse{
			Error: "Invalid JSON format",
		})
		return
	}

	missing, err := api.pathServ.GetMissingHashes(name, incomingHashes)
	if err != nil {
		zap.S().Errorf("Failed to query missing hashes: %v", err)
		c.AbortWithStatusJSON(500, model.ErrorResponse{
			Error: "failed to query missing hashes",
		})
		return
	}

	if missing == nil {
		missing = []string{}
	}

	c.JSON(200, missing)
}

// createNar godoc
//
//	@Summary		Create multipart NAR upload
//	@Description	Initialize a multipart upload for a NAR file.
//	@Tags			cache
//	@Produce		json
//	@Param			name		path		string	true	"Cache Name"
//	@Param			compression	query		string	true	"Compression method (xz or zst)"
//	@Success		200			{object}	map[string]string
//	@Failure		400			{object}	model.ErrorResponse
//	@Failure		500			{object}	model.ErrorResponse
//	@Router			/cache/{name}/multipart-nar [post]
func (api *cacheApi) createNar(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Creating multipart NAR for cache '%s'", name)

	compression := c.Query("compression")
	if compression != "xz" && compression != "zst" {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid compression",
		})
		return
	}

	id := uuid.New().String()
	filename := fmt.Sprintf("%s.nar.%s", id, compression)

	err := api.storage.CreatFile(name, filename)
	if err != nil {
		zap.S().Errorf("Failed to create multipart placeholder: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to create multipart placeholder",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"narId":    id,
		"uploadId": id,
	})
}

// redirect godoc
//
//	@Summary		Redirect to upload URL
//	@Description	Get the direct upload URL for a multipart NAR upload.
//	@Tags			cache
//	@Produce		json
//	@Param			name	path		string	true	"Cache Name"
//	@Param			narUuid	path		string	true	"NAR UUID"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		404		{object}	model.ErrorResponse
//	@Router			/cache/{name}/multipart-nar/{narUuid} [post]
func (api *cacheApi) redirect(c *gin.Context) {
	name := c.Param("name")
	narId := c.Param("narUuid")
	if narId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "missing narId",
		})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	host := c.Request.Host
	// We point back to the Management API path: /api/v1/cache/:name/multipart-nar/:narUuid
	uploadUrl := fmt.Sprintf("%s://%s/api/v1/cache/%s/multipart-nar/%s", scheme, host, name, narId)

	c.JSON(http.StatusOK, gin.H{
		"uploadUrl": uploadUrl,
	})
}

// uploadNarData godoc
//
//	@Summary		Upload NAR data
//	@Description	Upload raw NAR data for a given UUID in a cache.
//	@Tags			cache
//	@Accept			octet-stream
//	@Param			name	path	string	true	"Cache Name"
//	@Param			narUuid	path	string	true	"NAR UUID"
//	@Success		201
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Router			/cache/{name}/multipart-nar/{narUuid} [put]
func (api *cacheApi) uploadNarData(c *gin.Context) {
	name := c.Param("name")
	narUuid := c.Param("narUuid")
	if narUuid == "" {
		c.AbortWithStatusJSON(400, model.ErrorResponse{
			Error: "missing narUuid",
		})
		return
	}

	err := api.storage.WriteFile(name, narUuid, c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(500, model.ErrorResponse{
			Error: "failed to save to storage",
		})
		return
	}

	c.Header("Content-Location", "/")
	c.AbortWithStatus(http.StatusCreated)
}

type NarInfoCreate struct {
	CStoreHash   string   `json:"cStoreHash"`
	CStoreSuffix string   `json:"cStoreSuffix"`
	CNarHash     string   `json:"cNarHash"`
	CNarSize     int64    `json:"cNarSize"`
	CFileHash    string   `json:"cFileHash"`
	CFileSize    int64    `json:"cFileSize"`
	CReferences  []string `json:"cReferences"`
	CDeriver     string   `json:"cDeriver"`
	CSig         string   `json:"cSig"`
}

type CompletedMultipartUpload struct {
	NarInfoCreate NarInfoCreate `json:"narInfoCreate"`
}

// completeNar godoc
//
//	@Summary		Complete multipart NAR upload
//	@Description	Finalize a multipart upload and create the store path entry.
//	@Tags			cache
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string						true	"Cache Name"
//	@Param			narUuid	path		string						true	"NAR UUID"
//	@Param			request	body		CompletedMultipartUpload	true	"Completion details"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Router			/cache/{name}/multipart-nar/{narUuid}/complete [post]
func (api *cacheApi) completeNar(c *gin.Context) {
	name := c.Param("name")
	narUuid := c.Param("narUuid")
	zap.S().Infof("Completing multipart NAR upload: %s/%s", name, narUuid)

	var req CompletedMultipartUpload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	// Map to model
	sp := model.StorePath{
		StoreHash:   req.NarInfoCreate.CStoreHash,
		StoreSuffix: req.NarInfoCreate.CStoreSuffix,
		NarHash:     req.NarInfoCreate.CNarHash,
		NarSize:     req.NarInfoCreate.CNarSize,
		FileHash:    narUuid, // Using the UUID as the file identifier in storage
		FileSize:    req.NarInfoCreate.CFileSize,
		Deriver:     req.NarInfoCreate.CDeriver,
		References:  strings.Join(req.NarInfoCreate.CReferences, " "),
	}

	_, err := api.pathServ.Create(name, sp)
	if err != nil {
		zap.S().Errorf("Failed to finalize store path: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "failed to save store path",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// abortNar godoc
//
//	@Summary		Abort multipart NAR upload
//	@Description	Abort a multipart upload and clean up placeholder files.
//	@Tags			cache
//	@Param			name	path		string	true	"Cache Name"
//	@Param			narUuid	path		string	true	"NAR UUID"
//	@Success		200		{object}	map[string]interface{}
//	@Router			/cache/{name}/multipart-nar/{narUuid}/abort [post]
func (api *cacheApi) abortNar(c *gin.Context) {
	name := c.Param("name")
	narUuid := c.Param("narUuid")
	zap.S().Infof("Aborting multipart NAR upload: %s/%s", name, narUuid)

	// Clean up the placeholder file
	err := api.storage.DeleteFile(name, narUuid)
	if err != nil {
		zap.S().Warnf("Failed to clean up aborted NAR file: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{})
}
