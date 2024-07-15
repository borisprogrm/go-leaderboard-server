//go:build test
// +build test

package config

import (
	cacheprovider "go-leaderboard-server/internal/cache"
	cache_simple_provider "go-leaderboard-server/internal/cache/simple"
	dbprovider "go-leaderboard-server/internal/db"
	"os"

	db_inmemory_provider "go-leaderboard-server/internal/db/inmemory"
)

func init() {
	config = Config{
		IsDebug: true,
		IsTest:  true,
		Port:    os.Getenv("APP_PORT"),
		Db: DbConfig{
			Type: DBTYPE_INMEMORY,
			Config: &db_inmemory_provider.DbInMemoryProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: true,
				},
			},
		},
		Cache: CacheConfig{
			Type: CACHETYPE_SIMPLE,
			Config: &cache_simple_provider.CacheSimpleProviderConfig{
				CacheProviderBaseConfig: cacheprovider.CacheProviderBaseConfig{
					IsDebug: true,
					Ttl:     1000,
				},
			},
		},
		ApiUI: false,
	}
}
