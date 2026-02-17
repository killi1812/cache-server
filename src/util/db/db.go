// Package db contains connections for database
package db

import (
	"regexp"
	"time"

	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newPostgresConn(dsn string) *gorm.DB {
	// TODO: implement connection string

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func newSqliteConn(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

// New creates a new connection to the dabase based on the config
func New() *gorm.DB {
	var db *gorm.DB

	dsn := config.Config.CacheServer.Database
	switch GetDatabseType(dsn) {
	case Postgres:
		db = newPostgresConn(dsn)
	case Sqlite:
		db = newSqliteConn(dsn)
	default:
		db = newSqliteConn("file:db?mode=memory&cache=shared")
		// TODO: add log and paninc in prod
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Panicf("failed to get database connection: %+v", err)
	}

	// TODO: set as options
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
	}
	return db
}

var (

	// The DSN Key-Value format: Requires at least TWO recognized postgres keys
	// (host, dbname, user, sslmode, port, password) followed by an equals sign.
	pgStrictRegex = regexp.MustCompile(`(?i)^(?:(?:^|\s+)(?:host|dbname|user|sslmode|port|password)=\S+){6,}`)

	// sqliteRegex matches .db files, file: URIs, or in-memory strings
	sqliteRegex = regexp.MustCompile(`\.db(\?.*)?$|:memory:|^file:`)
)

type DatabseType int

const (
	Sqlite DatabseType = iota
	Postgres
	BadDbType
)

// GetDatabseType check by regex types for matching db conn strings
func GetDatabseType(dsn string) DatabseType {
	if pgStrictRegex.MatchString(dsn) {
		return Postgres
	}

	if sqliteRegex.MatchString(dsn) {
		return Sqlite
	}

	return BadDbType
}
