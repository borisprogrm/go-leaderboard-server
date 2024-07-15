//go:build production
// +build production

package config

import (
	cacheprovider "go-leaderboard-server/internal/cache"
	cache_simple_provider "go-leaderboard-server/internal/cache/simple"
	dbprovider "go-leaderboard-server/internal/db"
	mongo_provider "go-leaderboard-server/internal/db/mongodb"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	config = Config{
		IsDebug: false,
		IsTest:  false,
		Port:    os.Getenv("APP_PORT"),
		Db: DbConfig{
			Type: DBTYPE_MONGO,
			Config: &mongo_provider.MongoProviderConfig{
				DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
					IsDebug: false,
				},
				Uri:     os.Getenv("MONGODB_URI"),
				Options: options.Client().SetServerSelectionTimeout(5 * time.Second),
			},
		},
		Cache: CacheConfig{
			Type: CACHETYPE_SIMPLE,
			Config: &cache_simple_provider.CacheSimpleProviderConfig{
				CacheProviderBaseConfig: cacheprovider.CacheProviderBaseConfig{
					IsDebug: false,
					Ttl:     5000,
				},
			},
		},
		TimeoutServicesInit:     5000,
		TimeoutServerClose:      10000,
		TimeoutServicesShutdown: 5000,
		ApiUI:                   false,
	}
}
