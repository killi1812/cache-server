package db

import (
	"time"

	"github.com/killi1812/go-cache-server/model"
	"github.com/killi1812/go-cache-server/util/gormzap"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newPostgresConn() *gorm.DB {
	// TODO: implement connection string

	db, err := gorm.Open(postgres.Open(""), &gorm.Config{
		Logger: gormzap.NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func newSqliteConn() *gorm.DB {
	// TODO: implement connection string

	db, err := gorm.Open(sqlite.Open("file:db?mode=memory&cache=shared"), &gorm.Config{
		Logger: gormzap.NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func New() *gorm.DB {
	var db *gorm.DB

	if false {
		db = newPostgresConn()
	} else {
		db = newSqliteConn()
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Panicf("failed to get database connection: %+v", err)
	}

	// TODO: seet as options
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
	}
	return db
}
