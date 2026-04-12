package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func TestServerConf_GetDatabseType(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		dsn  string
		want db.DatabseType
	}{
		{
			name: "Good Sqlite conn string",
			dsn:  "test.db",
			want: db.Sqlite,
		},
		{
			name: "Good Sqlite conn string in memory",
			dsn:  "file::memory:?cache=shared",
			want: db.Sqlite,
		},
		{
			name: "Good Sqlite conn string in memory",
			dsn:  "test.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)",
			want: db.Sqlite,
		},
		{
			name: "Good Postgress conn string",
			dsn:  "host=localhost user=admin password=123 dbname=test port=5332 sslmode=disable",
			want: db.Postgres,
		},
		{
			name: "Bad Postgress conn string missing host",
			dsn:  "user=admin password=123 dbname=test port=5332 sslmode=disable",
			want: db.BadDbType,
		},
		{
			name: "Bad Postgress conn string missing user",
			dsn:  "host=localhost password=123 dbname=test port=5332 sslmode=disable",
			want: db.BadDbType,
		},
		{
			name: "Bad Postgress conn string missing password",
			dsn:  "host=localhost user=admin dbname=test port=5332 sslmode=disable",
			want: db.BadDbType,
		},
		{
			name: "Bad Postgress conn string missing dbname",
			dsn:  "host=localhost user=admin password=123 port=5332 sslmode=disable",
			want: db.BadDbType,
		},
		{
			name: "Bad Postgress conn string missing port",
			dsn:  "host=localhost user=admin password=123 dbname=test sslmode=disable",
			want: db.BadDbType,
		},
		{
			name: "Bad Postgress conn string missing sslmode",
			dsn:  "host=localhost user=admin password=123 port=5332 dbname=test",
			want: db.BadDbType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.GetDatabseType(tt.dsn)
			assert.Equal(t, tt.want, got, "Returned db types should match")
		})
	}
}

func TestNewAndMigration(t *testing.T) {
	app.Test()
	config.Config = config.NewConfig()
	config.Config.CacheServer.Database = "file:" + t.Name() + "?mode=memory&cache=shared"

	database := db.New()
	assert.NotNil(t, database)

	err := db.Migration(database)
	assert.NoError(t, err)
}

func TestLogger(t *testing.T) {
	app.Test()
	// Since we can't easily test zap output without complex hooks,
	// we just ensure calling them doesn't panic.
	config.Config = config.NewConfig()
	config.Config.CacheServer.Database = "file:" + t.Name() + "?mode=memory&cache=shared"
	database := db.New()
	
	l := database.Config.Logger
	ctx := context.Background()

	l.Info(ctx, "test info %s", "arg")
	l.Warn(ctx, "test warn %s", "arg")
	l.Error(ctx, "test error %s", "arg")
	
	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, os.ErrNotExist)

	l.Trace(ctx, time.Now().Add(-1*time.Second), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	newLogger := l.LogMode(logger.Silent)
	assert.NotNil(t, newLogger)
}
