package config

import (
	"errors"
	cacheprovider "go-leaderboard-server/internal/cache"
	dbprovider "go-leaderboard-server/internal/db"
	"go-leaderboard-server/internal/utils"
	"strconv"
)

type Config struct {
	IsDebug                 bool        // Debug environment flag
	IsTest                  bool        // Test environment flag
	Host                    string      `default:"0.0.0.0"` // Server host
	Port                    string      `default:"8415"`    // Server port
	Db                      DbConfig    // DB Provider Configuration
	Cache                   CacheConfig // Cache Provider Configuration
	TimeoutServicesInit     uint32      // Server initialization timeout (ms)
	TimeoutServerClose      uint32      // Server shutdown timeout (ms)
	TimeoutServicesShutdown uint32      // Services shutdown timeout (ms)
	ApiUI                   bool        // Swagger UI startup flag
}

const (
	DBTYPE_INMEMORY = iota
	DBTYPE_MONGO
	DBTYPE_POSTGRESQL
	DBTYPE_MYSQL
)

type DbConfig struct {
	Type   int
	Config dbprovider.IDBProviderConfig
}

const (
	CACHETYPE_SIMPLE = iota
)

type CacheConfig struct {
	Type   int
	Config cacheprovider.ICacheProviderConfig
}

var config Config
var configPtr *Config

func configApplyAndValidate(c *Config) {
	var err error

	err = utils.ApplyDefaults(c)
	if err != nil {
		panic(err)
	}

	port := c.Port
	_, e := strconv.ParseUint(port, 10, 16)
	if e != nil {
		err = errors.Join(err, errors.New("wrong port value"))
	}

	if err != nil {
		panic(err)
	}
}

func GetAppConfig() *Config {
	if configPtr == nil {
		configPtr = &config
		configApplyAndValidate(configPtr)
	}
	return configPtr
}
