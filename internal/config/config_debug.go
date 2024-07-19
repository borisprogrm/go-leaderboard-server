//go:build debug
// +build debug

package config

import (
	cacheprovider "go-leaderboard-server/internal/cache"
	cache_simple_provider "go-leaderboard-server/internal/cache/simple"
	dbprovider "go-leaderboard-server/internal/db"
	"os"

	db_inmemory_provider "go-leaderboard-server/internal/db/inmemory"
	//~ redis_provider "go-leaderboard-server/internal/db/redis"
	//~ mongo_provider "go-leaderboard-server/internal/db/mongodb"
	//~ postgre_provider "go-leaderboard-server/internal/db/postgresql"
	//~ mysql_provider "go-leaderboard-server/internal/db/mysql"
)

func init() {
	config = Config{
		IsDebug: true,
		IsTest:  false,
		Port:    os.Getenv("APP_PORT"),
		Db: DbConfig{
			Type: DBTYPE_INMEMORY,
			Config: &db_inmemory_provider.DbInMemoryProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: true,
				},
			},
		},
		/* OR
		Db: DbConfig{
			Type: DBTYPE_REDIS,
			Config: &redis_provider.RedisProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: true,
				},
				Opts: redis_provider.RedisOptions{
					Addr: "localhost:6379",
				},
			},
		},
		*/
		/* OR
		Db: DbConfig{
			Type: DBTYPE_MONGO,
			Config: &mongo_provider.MongoProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: true,
				},
				Uri: "mongodb://localhost:27017"
			},
		},
		*/
		/* OR
		Db: DbConfig{
			Type: DBTYPE_POSTGRESQL,
			Config: &postgre_provider.PostgreProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: true,
				},
				ConnStr: "postgres://admin:admpass@localhost:5432/GoLeaderboard",
			},
		},
		*/
		/* OR
		Db: DbConfig{
			Type: DBTYPE_MYSQL,
			Config: &mysql_provider.MySqlProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: true,
				},
				ConnStr: "admin:admpass@tcp(localhost:3306)/GoLeaderboard",
			},
		},
		*/
		Cache: CacheConfig{
			Type: CACHETYPE_SIMPLE,
			Config: &cache_simple_provider.CacheSimpleProviderConfig{
				CacheProviderBaseConfig: cacheprovider.CacheProviderBaseConfig{
					IsDebug: true,
					Ttl:     1000,
				},
			},
		},
		TimeoutServicesInit:     5000,
		TimeoutServerClose:      10000,
		TimeoutServicesShutdown: 5000,
		ApiUI:                   true,
	}
}
