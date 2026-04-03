package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/auth"
	"go.uber.org/zap"
)

func (api *cacheApi) name(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to read cache '%s'", name)

	cache, err := api.cacheServ.Read(name)
	if err != nil {
		zap.S().Errorf("Failed to read cache '%s', err: %v", name, err)
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to read cache"})
		return
	}

	if cache.Access == model.Private {
		// TODO: protect
		// not like this this is middleware only
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

func (api *cacheApi) narinfo(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to retrive missing narinfo '%s'", name)

	var incomingHashes []string

	if err := c.ShouldBindJSON(&incomingHashes); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	missing, err := api.pathServ.GetMissingHashes(name, incomingHashes)
	if err != nil {
		zap.S().Errorf("Failed to query missing hashes: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, missing)
}

func (api *cacheApi) createNar(c *gin.Context) {
	name := c.Param("name")
	zap.S().Infof("Trying to retrive missing narinfo '%s'", name)

	compression := c.Query("compression")
	if compression != "xz" && compression != "zst" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid compression"})
		return
	}

	id := uuid.New().String()
	filename := fmt.Sprintf("%s.nar.%s", id, compression)

	err := api.storage.CreatFile(name, filename)
	if err != nil {
		zap.S().Errorf("Failed to create multipart placeholder: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 4. Return the IDs
	// Gin handles the Content-Type and Content-Length headers
	c.JSON(http.StatusOK, gin.H{
		"narId":    id,
		"uploadId": id,
	})
}

func (api *cacheApi) redirect(c *gin.Context) {
	narId := c.Param("narUuid")
	if narId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing narId"})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	uploadUrl := fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, narId)

	// TODO: add check if path or cache exists

	c.JSON(http.StatusOK, gin.H{
		"uploadUrl": uploadUrl,
	})
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
	// Parts []any `json:"parts"` // Ignored for now
}

func (api *cacheApi) completeNar(c *gin.Context) {
	name := c.Param("name")
	narUuid := c.Param("narUuid")
	zap.S().Infof("Completing multipart NAR upload: %s/%s", name, narUuid)

	var req CompletedMultipartUpload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save store path"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (api *cacheApi) abortNar(c *gin.Context) {
	name := c.Param("name")
	narUuid := c.Param("narUuid")
	zap.S().Infof("Aborting multipart NAR upload: %s/%s", name, narUuid)

	// Clean up the placeholder file
	err := api.storage.DeleteFile(name + "/" + narUuid) // Assuming directory structure
	if err != nil {
		zap.S().Warnf("Failed to clean up aborted NAR file: %v", err)
	}

	c.Status(http.StatusNoContent)
}
