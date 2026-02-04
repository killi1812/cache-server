package config

import (
	"go.uber.org/zap"
	"gopkg.in/ini.v1"
)

/*
[cache-server]
hostname = cache-server
cache-dir = binary-caches
server-port = 12345
database = dbfile.db
deploy-port = 54321
key = secret
*/

var ConfigPath string

var Config *ServerConf

type AppConfig struct {
	CacheServer *ServerConf `ini:"cache-server"`
}

type ServerConf struct {
	Hostname   string `ini:"hostname"`
	CacheDir   string `ini:"cache-dir"`
	ServerPort int    `ini:"server-port"`
	// connections string or sqlite file name
	Database   string `ini:"database"`
	DeployPort int    `ini:"deploy-port"`
	Key        string `ini:"key"`
}

func NewConfig() *AppConfig {
	c := &AppConfig{CacheServer: &ServerConf{}}
	c.CacheServer.Hostname = "cache-server"
	c.CacheServer.CacheDir = "binary-caches"
	c.CacheServer.ServerPort = 12345
	c.CacheServer.Database = "dbfile.db"
	c.CacheServer.DeployPort = 12345
	c.CacheServer.Key = "secret"
	return c
}

func ReadConfig() error {
	zap.S().Debugf("Reading config")

	// Create a new config object
	config := NewConfig()

	cfg, err := ini.Load(ConfigPath)
	if err != nil {

		zap.S().Errorf("Failed to read config using defaults")
		zap.S().Error(err.Error())
		zap.S().Debugf("Default config: %+v", *config.CacheServer)

		Config = config.CacheServer
		return err
	}

	// Map config file values
	cfg.MapTo(config)
	Config = config.CacheServer

	zap.S().Debugf("Config read successfully")
	zap.S().Debugf("Config: %+v", *config.CacheServer)
	return nil
}
