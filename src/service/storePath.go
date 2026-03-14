package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/objstor"
	"gorm.io/gorm"
)

type StorePathSrv struct {
	db    *gorm.DB
	store objstor.ObjectStorage
}

func NewStorePathSrv() *StorePathSrv {
	var srv *StorePathSrv

	app.Invoke(func(db *gorm.DB, store objstor.ObjectStorage) {
		db = db.
			Table("store_paths").
			Joins("JOIN binary_caches ON binary_caches.id = store_paths.binary_cache_id")
		srv = &StorePathSrv{db, store}
	})

	return srv
}

// ReadAll fetches all store paths for a specific cache
func (s *StorePathSrv) ReadAll(cacheName string) ([]model.StorePath, error) {
	var paths []model.StorePath
	result := s.db.
		Where("binary_caches.name = ?", cacheName).
		Find(&paths)
	return paths, result.Error
}

// Read fetches a specific store path by its hash and cache name
func (s *StorePathSrv) Read(storeHash string, cache string) (*model.StorePath, error) {
	var path model.StorePath
	result := s.db.Where("store_paths.store_hash = ? AND binary_caches.name = ?", storeHash, cache).First(&path)

	if result.Error != nil {
		// if result.Error == gorm.ErrRecordNotFound {
		// 	return nil, nil
		// }
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

	// 3. Delete the actual NAR file from Object Storage (S3/Local)
	// Note: You'll need to determine the extension (e.g., .nar.xz)
	// based on your storage logic.
	return s.store.DeleteFile(path.FileHash)
}

// TODO: check wtf is this needed for
func (s *StorePathSrv) GenerateNarInfo(p *model.StorePath, privateKey string) (string, error) {
	// Fingerprint: 1;/nix/store/hash-suffix;narhash;narsize;refs
	refs := strings.Split(p.References, " ")
	for i, r := range refs {
		refs[i] = "/nix/store/" + r
	}
	fingerprint := fmt.Sprintf("1;/nix/store/%s-%s;%s;%d;%s",
		p.StoreHash, p.StoreSuffix, p.NarHash, p.NarSize, strings.Join(refs, ","))

	// Signing logic
	parts := strings.Split(privateKey, ":")
	seed, _ := base64.StdEncoding.DecodeString(parts[1])
	privKey := ed25519.NewKeyFromSeed(seed)
	sig := ed25519.Sign(privKey, []byte(fingerprint))

	sigString := fmt.Sprintf("%s:%s", parts[0], base64.StdEncoding.EncodeToString(sig))

	return fmt.Sprintf(`StorePath: /nix/store/%s-%s
URL: nar/%s.nar
FileHash: sha256:%s
FileSize: %d
NarHash: %s
NarSize: %d
Deriver: %s
System: x86_64-linux
References: %s
Sig: %s
`, p.StoreHash, p.StoreSuffix, p.FileHash, p.FileHash, p.FileSize, p.NarHash, p.NarSize, p.Deriver, p.References, sigString), nil
}

func (s *StorePathSrv) GetMissingHashes(cacheName string, incomingHashes []string) ([]string, error) {
	if len(incomingHashes) == 0 {
		return []string{}, nil
	}

	var foundHashes []string

	// 1. Find which of the INCOMING hashes actually exist in the DB
	err := s.db.Model(&model.StorePath{}).
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
