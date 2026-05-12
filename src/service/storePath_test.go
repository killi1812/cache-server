package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/killi1812/go-cache-server/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateNarInfo(t *testing.T) {
	srv := &StorePathSrv{}

	// Sample data from user's "good" example
	p := &model.StorePath{
		StoreHash:   "ffgmyxfrc3v77azm9g8lix2kp3rcf443",
		StoreSuffix: "testhello",
		NarHash:     "sha256:1p4a6kwhz5h1ppcqc5k10mgjbbqj55pzwr98d68n048yrqs3bj5s",
		NarSize:     191640,
		FileHash:    "sha256:6d4a231752b5cebfe0f73466997e8c20be47a014e05b6b37e89c1e5465f97920",
		FileSize:    37364,
		References:  "ffgmyxfrc3v77azm9g8lix2kp3rcf443-testhello j193mfi0f921y0kfs8vjc1znnr45ispv-glibc-2.40-66",
		Deriver:     "zcxchykc7js9mcb4nq58283sddh5qr48-testhello.drv",
	}

	seed := make([]byte, 32)
	ed25519.NewKeyFromSeed(seed)
	privKeyStr := "test.localhost-1:" + base64.StdEncoding.EncodeToString(seed)

	output, err := srv.GenerateNarInfo(p, privKeyStr)
	assert.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedFields := []string{
		"StorePath: /nix/store/ffgmyxfrc3v77azm9g8lix2kp3rcf443-testhello",
		"URL: nar/6d4a231752b5cebfe0f73466997e8c20be47a014e05b6b37e89c1e5465f97920.nar.xz",
		"Compression: xz",
		"FileHash: sha256:6d4a231752b5cebfe0f73466997e8c20be47a014e05b6b37e89c1e5465f97920",
		"FileSize: 37364",
		"NarHash: sha256:1p4a6kwhz5h1ppcqc5k10mgjbbqj55pzwr98d68n048yrqs3bj5s",
		"NarSize: 191640",
		"References: ffgmyxfrc3v77azm9g8lix2kp3rcf443 j193mfi0f921y0kfs8vjc1znnr45ispv",
		"Deriver: zcxchykc7js9mcb4nq58283sddh5qr48-testhello.drv",
	}

	for i, expected := range expectedFields {
		assert.Equal(t, expected, lines[i], "Field mismatch at line %d", i+1)
	}

	assert.True(t, strings.HasPrefix(lines[len(lines)-1], "Sig: test.localhost-1:"), "Signature line should have correct prefix")
}
