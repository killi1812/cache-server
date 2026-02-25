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
		srv = &StorePathSrv{db, store}
	})

	return srv
}

// ReadAll fetches all store paths for a specific cache
func (s *StorePathSrv) ReadAll(cache string) ([]model.StorePath, error) {
	var paths []model.StorePath
	result := s.db.Where("cache_name = ?", cache).Find(&paths)
	return paths, result.Error
}

// Read fetches a specific store path by its hash and cache name
func (s *StorePathSrv) Read(storeHash string, cache string) (*model.StorePath, error) {
	var path model.StorePath
	result := s.db.Where("store_hash = ? AND cache_name = ?", storeHash, cache).First(&path)

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
