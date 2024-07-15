package db_inmemory_provider

import (
	"context"
	"errors"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"sort"
	"sync"
)

var logger = log.GetLogger()

type DbInMemoryProviderConfig struct {
	dbprovider.DBProviderBaseConfig
}

type DbInMemoryProvider struct {
	mutex sync.RWMutex
	data  map[string](map[string]dbprovider.UserProperties)
}

func NewDbInMemoryProvider() *DbInMemoryProvider {
	return &DbInMemoryProvider{
		data: make(map[string](map[string]dbprovider.UserProperties)),
	}
}

func (p *DbInMemoryProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	logger.Debug("DB provider initialization")

	conf, ok := config.(*DbInMemoryProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}

	if !conf.IsDebug {
		logger.Warn("DbInMemoryProvider should only be used for testing purposes only!")
	}

	return nil
}

func (p *DbInMemoryProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.data[gameId]; !ok {
		p.data[gameId] = make(map[string]dbprovider.UserProperties)
	}

	p.data[gameId][userId] = userProp

	return nil
}

func (p *DbInMemoryProvider) Delete(ctx context.Context, gameId string, userId string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.data[gameId]; ok {
		delete(p.data[gameId], userId)
	}

	return nil
}

func (p *DbInMemoryProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	gd, ok := p.data[gameId]
	if !ok {
		return nil, nil
	}

	ud, ok := gd[userId]
	if !ok {
		return nil, nil
	}

	return &ud, nil
}

func (p *DbInMemoryProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	gd, ok := p.data[gameId]
	if !ok {
		return dbprovider.TopData{}, nil
	}

	uscores := make([]dbprovider.UserData, 0, len(gd))
	for k, v := range gd {
		uscores = append(uscores,
			dbprovider.UserData{
				UserId:         k,
				UserProperties: dbprovider.UserProperties{Score: v.Score, Name: v.Name, Params: v.Params},
			},
		)
	}

	sort.Slice(uscores, func(i, j int) bool {
		return uscores[j].UserProperties.Score < uscores[i].UserProperties.Score
	})

	top := uscores[:min(len(uscores), int(nTop))]
	return top, nil
}

func (p *DbInMemoryProvider) Shutdown(ctx context.Context) error {
	logger.Debug("DB provider shutdown")

	/* do nothing */

	return nil
}
