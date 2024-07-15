package cacheprovider

import (
	"context"
	dbprovider "go-leaderboard-server/internal/db"
)

type CacheProviderBaseConfig struct {
	IsDebug bool   // Debug flag
	Ttl     uint32 // Cache element lifetime (ms)
}

func (c *CacheProviderBaseConfig) GetBaseConfig() *CacheProviderBaseConfig {
	return c
}

type ICacheProviderConfig interface {
	GetBaseConfig() *CacheProviderBaseConfig
}

type CacheData struct {
	Exp  int64
	Cnt  uint32
	Data dbprovider.TopData
}

type ICacheProvider interface {
	Initialize(ctx context.Context, config ICacheProviderConfig, dbProvider dbprovider.IDbProvider) error
	Top(gctx context.Context, ameId string, nTop uint32) (dbprovider.TopData, error)
	Shutdown(ctx context.Context) error
}
