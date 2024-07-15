package cache_simple_provider

import (
	"context"
	"errors"
	cacheprovider "go-leaderboard-server/internal/cache"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"go-leaderboard-server/internal/utils"
	"sync"

	"golang.org/x/sync/singleflight"
)

var logger = log.GetLogger()

type CacheSimpleProviderConfig struct {
	cacheprovider.CacheProviderBaseConfig
}

type CacheSimpleProvider struct {
	cache      map[string]*cacheprovider.CacheData
	dbprovider dbprovider.IDbProvider
	ttl        uint32
	mutex      sync.RWMutex
	sfg        singleflight.Group
	clock      *utils.IClock
}

func NewCacheSimpleProvider(clock *utils.IClock) *CacheSimpleProvider {
	return &CacheSimpleProvider{
		cache: make(map[string]*cacheprovider.CacheData),
		clock: clock,
	}
}

func (p *CacheSimpleProvider) Initialize(ctx context.Context, config cacheprovider.ICacheProviderConfig, dbProvider dbprovider.IDbProvider) error {
	logger.Debug("Cache provider initialization")

	conf, ok := config.(*CacheSimpleProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}

	p.ttl = conf.Ttl
	p.dbprovider = dbProvider

	return nil
}

func (p *CacheSimpleProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	if p.dbprovider == nil {
		return nil, errors.New("uninitialized")
	}

	now := (*(p.clock)).Now().UnixMilli()

	p.mutex.RLock()
	cacheData, ok := p.cache[gameId]
	if ok && cacheData.Exp > now && cacheData.Cnt >= nTop {
		topData := cacheData.Data
		p.mutex.RUnlock()
		if len(topData) > int(nTop) {
			topData = topData[:int(nTop)]
		}
		return topData, nil
	}
	p.mutex.RUnlock()

	topData, err, _ := p.sfg.Do(gameId, func() (any, error) {
		data, err := p.dbprovider.Top(ctx, gameId, nTop)
		if err == nil {
			p.mutex.Lock()
			p.cache[gameId] = &cacheprovider.CacheData{
				Exp:  now + int64(p.ttl),
				Cnt:  nTop,
				Data: data,
			}
			p.mutex.Unlock()
		}
		return data, err
	})

	return topData.(dbprovider.TopData), err
}

func (p *CacheSimpleProvider) Shutdown(ctx context.Context) error {
	logger.Debug("Cache provider shutdown")

	/* do nothing */

	return nil
}
