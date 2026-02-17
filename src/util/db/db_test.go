package db_test

import (
	"testing"

	"github.com/killi1812/go-cache-server/util/db"
	"github.com/stretchr/testify/assert"
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
