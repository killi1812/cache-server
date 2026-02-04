package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		configName string
		want       *AppConfig
		wantErr    error
	}{
		{
			name:       "Read good confg",
			configName: "testdata/good.conf",
			want:       NewConfig(),
			wantErr:    nil,
		},
		{
			name:       "Read bad confg",
			configName: "testdata/bad.conf",
			want:       NewConfig(),
			wantErr:    ErrBadConfig,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf, err := readConfig(tt.configName)
			if tt.wantErr != nil {
				assert.Error(t, err, "Err should not be nil")
				assert.ErrorIs(t, err, tt.wantErr, "Error should have ErrBadConfig")
			} else {
				assert.Nil(t, err, "No error should be returned")
				assert.NotNil(t, conf, "Counfiguration should not be nil")

				assert.Equal(t, tt.want.CacheServer.Hostname, conf.CacheServer.Hostname, "Hostname should match")
				assert.Equal(t, tt.want.CacheServer.CacheDir, conf.CacheServer.CacheDir, "CacheDir should match")
				assert.Equal(t, tt.want.CacheServer.Database, conf.CacheServer.Database, "Database should match")
				assert.Equal(t, tt.want.CacheServer.DeployPort, conf.CacheServer.DeployPort, "DeployPort should match")
				assert.Equal(t, tt.want.CacheServer.ServerPort, conf.CacheServer.ServerPort, "ServerPort should match")
				assert.Equal(t, tt.want.CacheServer.Key, conf.CacheServer.Key, "Key should match")

				// TODO: add asserts
			}
		})
	}
}

var loadConfigMutex = sync.Mutex{}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		configName string
		want       *AppConfig
		wantErr    error
	}{
		{
			name:       "Read good confg",
			configName: "testdata/good.conf",
			want:       NewConfig(),
			wantErr:    nil,
		},
		{
			name:       "Read bad confg",
			configName: "testdata/bad.conf",
			want:       NewConfig(),
			wantErr:    ErrBadConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadConfigMutex.Lock()
			defer loadConfigMutex.Unlock()

			Config = nil
			ConfigPath = tt.configName

			err := LoadConfig()
			if tt.wantErr != nil {
				assert.Error(t, err, "Err should not be nil")
			} else {
				assert.Nil(t, err, "No error should be returned")
			}
		})
	}
}
