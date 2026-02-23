// Package db contains connections for database
package db

import (
	"regexp"
	"strings"
	"time"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newPostgresConn(dsn string) *gorm.DB {
	zap.S().Infof("Opening new postgress connection")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func newSqliteConn(dsn string) *gorm.DB {
	zap.S().Infof("Opening new sqlite connection")

	// BUG : not cascading and not setting foreign_keys to true
	if !strings.Contains(dsn, "_foreign_keys") || !strings.Contains(dsn, "_fk") {
		zap.S().Warn("Sqlite connection missing foreign_keys option")
		zap.S().Infof("Adding foreign_keys option, dsn: %s", dsn)
		if strings.Contains(dsn, "?") {
			dsn = "file:" + dsn + "&foreign_keys=true"
		} else {
			dsn = "file:" + dsn + "?foreign_keys=true"
		}
		zap.S().Infof("Added foreign_keys option, dsn: %s", dsn)
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func Migration(db *gorm.DB) error {
	zap.S().Infof("Trying to run AutoMigrate")
	if err := db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().DPanicf("Can't run AutoMigrate err = %+v", err)
		return err
	}
	zap.S().Infof("Successfull auto migrate")
	return nil
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
		if app.Build == app.BuildProd {
			zap.S().Panicf("failed to identify database connection string")
			return nil // just to be sure
		}
		zap.S().Warnf("Using in memory database")
		db = newSqliteConn("file:db?mode=memory&cache=shared")
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Panicf("failed to get database connection: %+v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
