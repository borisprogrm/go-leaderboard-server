//go:build debug
// +build debug

package config

import (
	cacheprovider "go-leaderboard-server/internal/cache"
	cache_simple_provider "go-leaderboard-server/internal/cache/simple"
	dbprovider "go-leaderboard-server/internal/db"
	"os"

	db_inmemory_provider "go-leaderboard-server/internal/db/inmemory"
	//~ mongo_provider "go-leaderboard-server/internal/db/mongodb"
	//~ postgre_provider "go-leaderboard-server/internal/db/postgresql"
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
			Type: DBTYPE_MONGO,
			Config: &mongo_provider.MongoProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: false,
				},
				Uri:     "mongodb://localhost:27017"
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
