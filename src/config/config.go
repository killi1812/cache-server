// Package config contains the apps runtime config
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/ini.v1"
)

var (
	ConfigPath   string
	Config       *AppConfig
	ErrBadConfig = errors.New("bad config")
)

type (
	AppConfig struct {
		CacheServer *ServerConf `ini:"cache-server"`
		Minio       *MinioConf  `ini:"minio"`
	}

	MinioConf struct {
		Endpoint   string `ini:"endpoint"`
		CredID     string `ini:"id"`
		CredSecret string `ini:"secret"`
		CredToken  string `ini:"token,omitempty"`
		UseSSL     bool   `ini:"use-ssl"`
	}

	ServerConf struct {
		Hostname   string `ini:"hostname"`
		CacheDir   string `ini:"cache-dir"`
		ServerPort int    `ini:"server-port"`
		// connections string or sqlite file name
		Database   string `ini:"database"`
		DeployPort int    `ini:"deploy-port"`
		Key        string `ini:"key"`
	}
)

func NewConfig() *AppConfig {
	c := &AppConfig{
		CacheServer: &ServerConf{
			Hostname:   "cache-server",
			CacheDir:   "binary-caches",
			ServerPort: 12345,
			Database:   "dbfile.db",
			DeployPort: 54321,
			Key:        "secret",
		},
		Minio: &MinioConf{
			Endpoint:   "play.min.io",
			CredID:     "Q3AM3UQ867SPQQA43P2F",
			CredSecret: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
			CredToken:  "",
			UseSSL:     true,
		},
	}

	return c
}

// LoadConfig will try read config
//
// if it fails: return defaults and error
func LoadConfig() error {
	zap.S().Debugf("Reading config")

	config, err := readConfig(ConfigPath)
	tmpb := strings.Builder{}
	tmpe := json.NewEncoder(&tmpb)
	tmpe.SetIndent("", "   ")
	tmpe.Encode(config)

	if err != nil {
		zap.S().Warn("Failed to load config using defaults")
		zap.S().Debug(err.Error())

		zap.S().Debugf("Default config: %s", tmpb.String())

		Config = config
		return err
	}
	zap.S().Debugf("Config read successfully")
	zap.S().Debugf("Default config: %s", tmpb.String())

	// set global config
	Config = config
	return nil
}

func readConfig(filename string) (*AppConfig, error) {
	// Create a new config object
	config := NewConfig()

	cfg, err := ini.Load(filename)
	if err != nil {
		return config, fmt.Errorf("%s: %w", ErrBadConfig, err)
	}

	err = cfg.MapTo(config)
	if err != nil {
		return config, nil
	}

	return config, nil
}
