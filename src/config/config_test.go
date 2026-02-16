package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// BUG: tests seem not to work propperly
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
			want:       nil,
			wantErr:    ErrBadConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf, err := readConfig(tt.configName)
			if tt.wantErr != nil {
				assert.Error(t, err, "Err should not be nil")
				assert.ErrorContains(t, err, tt.wantErr.Error(), "Error should have ErrBadConfig")
			} else {
				assert.Nil(t, err, "No error should be returned")
				assert.NotNil(t, conf, "Counfiguration should not be nil")

				// CacheServer asserts
				assert.Equal(t, tt.want.CacheServer.Hostname, conf.CacheServer.Hostname, "Hostname should match")
				assert.Equal(t, tt.want.CacheServer.CacheDir, conf.CacheServer.CacheDir, "CacheDir should match")
				assert.Equal(t, tt.want.CacheServer.Database, conf.CacheServer.Database, "Database should match")
				assert.Equal(t, tt.want.CacheServer.DeployPort, conf.CacheServer.DeployPort, "DeployPort should match")
				assert.Equal(t, tt.want.CacheServer.ServerPort, conf.CacheServer.ServerPort, "ServerPort should match")
				assert.Equal(t, tt.want.CacheServer.Key, conf.CacheServer.Key, "Key should match")

				// Minio asserts
				assert.Equal(t, tt.want.Minio.Endpoint, conf.Minio.Endpoint, "Endpoint should match")
				assert.Equal(t, tt.want.Minio.CredID, conf.Minio.CredID, "Id should match")
				assert.Equal(t, tt.want.Minio.CredSecret, conf.Minio.CredSecret, "Secret should match")
				assert.Equal(t, tt.want.Minio.CredToken, conf.Minio.CredToken, "Token should match")
				assert.Equal(t, tt.want.Minio.UseSSL, conf.Minio.UseSSL, "Use-ssl should match")

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

func TestServerConf_GetDatabseType(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		conf ServerConf
		want DatabseType
	}{
		{
			name: "Good Sqlite conn string",
			conf: ServerConf{Database: "test.db"},
			want: Sqlite,
		},
		{
			name: "Good Sqlite conn string in memory",
			conf: ServerConf{Database: "file::memory:?cache=shared"},
			want: Sqlite,
		},
		{
			name: "Good Sqlite conn string in memory",
			conf: ServerConf{Database: "test.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)"},
			want: Sqlite,
		},
		{
			name: "Good Postgress conn string",
			conf: ServerConf{Database: "host=localhost user=admin password=123 dbname=test port=5332 sslmode=disable"},
			want: Postgres,
		},
		{
			name: "Bad Postgress conn string missing host",
			conf: ServerConf{Database: "user=admin password=123 dbname=test port=5332 sslmode=disable"},
			want: BadDbType,
		},
		{
			name: "Bad Postgress conn string missing user",
			conf: ServerConf{Database: "host=localhost password=123 dbname=test port=5332 sslmode=disable"},
			want: BadDbType,
		},
		{
			name: "Bad Postgress conn string missing password",
			conf: ServerConf{Database: "host=localhost user=admin dbname=test port=5332 sslmode=disable"},
			want: BadDbType,
		},
		{
			name: "Bad Postgress conn string missing dbname",
			conf: ServerConf{Database: "host=localhost user=admin password=123 port=5332 sslmode=disable"},
			want: BadDbType,
		},
		{
			name: "Bad Postgress conn string missing port",
			conf: ServerConf{Database: "host=localhost user=admin password=123 dbname=test sslmode=disable"},
			want: BadDbType,
		},
		{
			name: "Bad Postgress conn string missing sslmode",
			conf: ServerConf{Database: "host=localhost user=admin password=123 port=5332 dbname=test"},
			want: BadDbType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ServerConf.GetDatabseType(tt.conf)
			assert.Equal(t, tt.want, got, "Returned db types should match")
		})
	}
}
