package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/objstor"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StorePathSrv struct {
	db    *gorm.DB
	store objstor.ObjectStorage
}

func NewStorePathSrv(db *gorm.DB, store objstor.ObjectStorage) *StorePathSrv {
	return &StorePathSrv{db, store}
}

func (s *StorePathSrv) dbWithCache() *gorm.DB {
	return s.db.Session(&gorm.Session{}).
		Table("store_paths").
		Joins("JOIN binary_caches ON binary_caches.id = store_paths.binary_cache_id")
}

// ReadAll fetches all store paths for a specific cache
func (s *StorePathSrv) ReadAll(cacheName string) ([]model.StorePath, error) {
	var paths []model.StorePath
	result := s.dbWithCache().
		Where("binary_caches.name = ?", cacheName).
		Find(&paths)
	return paths, result.Error
}

// Read fetches a specific store path by its hash and cache name
func (s *StorePathSrv) Read(storeHash string, cache string) (*model.StorePath, error) {
	var path model.StorePath
	result := s.dbWithCache().Where("store_paths.store_hash = ? AND binary_caches.name = ?", storeHash, cache).First(&path)

	if result.Error != nil {
		return nil, result.Error
	}

	return &path, nil
}

// Delete removes the database record and potentially the associated file
func (s *StorePathSrv) Delete(storeHash string, cache string) error {
	// 1. Find the record first to get the FileHash (needed to delete from ObjectStorage)
	path, err := s.Read(storeHash, cache)
	if err != nil {
		return err
	}
	if path == nil {
		return fmt.Errorf("store path not found")
	}

	err = s.db.Delete(path).Error
	if err != nil {
		return err
	}

	return s.store.DeleteFile(cache, path.FileHash)
}

func (s *StorePathSrv) GenerateNarInfo(p *model.StorePath, privateKey string) (string, error) {
	// Fingerprint: 1;StorePath;NarHash;NarSize;References
	var refsList []string
	if p.References != "" {
		refsList = strings.Fields(p.References)
	}
	zap.S().Debugf("References: %v", refsList)

	fullPaths := make([]string, len(refsList))
	for i, r := range refsList {
		fullPaths[i] = "/nix/store/" + r
	}
	zap.S().Debugf("Full paths : %v", fullPaths)

	cleanHash := strings.TrimPrefix(p.NarHash, "sha256:")
	zap.S().Debugf("CleanNarHash: %s", cleanHash)
	fingerprint := fmt.Sprintf("1;/nix/store/%s-%s;%s;%d;%s",
		p.StoreHash, p.StoreSuffix, cleanHash, p.NarSize, strings.Join(fullPaths, ","))
	zap.S().Debugf("Fingerprint : %v", fingerprint)

	// Signing logic
	parts := strings.Split(privateKey, ":")
	keyString, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		zap.S().Errorf("Failed to decode string")
		return "", err
	}
	privKey := ed25519.PrivateKey(keyString)
	sig := ed25519.Sign(privKey, []byte(fingerprint))

	keyName := parts[0]
	sigString := fmt.Sprintf("%s:%s", keyName, base64.StdEncoding.EncodeToString(sig))

	// Compression detection and URL
	compression := "none"
	url := fmt.Sprintf("nar/%s.nar", p.FileHash)

	// If the file in DB was recorded with an extension, or we have size difference
	if p.FileSize < p.NarSize {
		compression = "xz"
		url += ".xz"
	}

	res := fmt.Sprintf(`StorePath: /nix/store/%s-%s
URL: %s
Compression: %s
FileHash: sha256:%s
FileSize: %d
NarHash: %s
NarSize: %d
References: %s
Deriver: %s
Sig: %s
`, p.StoreHash, p.StoreSuffix, url, compression, p.FileHash, p.FileSize, p.NarHash, p.NarSize, p.References, p.Deriver, sigString)

	zap.S().Debugf("Generated NarInfo for %s:\n%s", p.StoreHash, res)
	return res, nil
}

func (s *StorePathSrv) GetMissingHashes(cacheName string, incomingHashes []string) ([]string, error) {
	if len(incomingHashes) == 0 {
		return []string{}, nil
	}

	var foundHashes []string

	// 1. Find which of the INCOMING hashes actually exist in the DB
	err := s.dbWithCache().
		Where("binary_caches.name = ? AND store_paths.store_hash IN ?", cacheName, incomingHashes).
		Pluck("store_paths.store_hash", &foundHashes).Error
	if err != nil {
		return nil, err
	}

	// 2. Convert found hashes to a map for quick diffing
	foundMap := make(map[string]struct{}, len(foundHashes))
	for _, h := range foundHashes {
		foundMap[h] = struct{}{}
	}

	// 3. The 'missing' ones are the ones we asked for but didn't find
	var missing []string
	for _, h := range incomingHashes {
		if _, found := foundMap[h]; !found {
			missing = append(missing, h)
		}
	}

	return missing, nil
}

func (s *StorePathSrv) Create(cacheName string, path model.StorePath) (*model.StorePath, error) {
	zap.S().Infof("Creating store path %s in cache %s", path.StoreHash, cacheName)

	var cache model.BinaryCache
	err := s.db.Session(&gorm.Session{}).Where("name = ?", cacheName).First(&cache).Error
	if err != nil {
		return nil, err
	}

	path.BinaryCacheId = cache.ID
	err = s.db.Session(&gorm.Session{}).Create(&path).Error
	if err != nil {
		return nil, err
	}

	return &path, nil
}
