package services

import (
	"context"
	"errors"
	cacheprovider "go-leaderboard-server/internal/cache"
	cache_simple_provider "go-leaderboard-server/internal/cache/simple"

	"go-leaderboard-server/internal/config"
	dbprovider "go-leaderboard-server/internal/db"

	db_inmemory_provider "go-leaderboard-server/internal/db/inmemory"
	mongo_provider "go-leaderboard-server/internal/db/mongodb"
	mysql_provider "go-leaderboard-server/internal/db/mysql"
	postgre_provider "go-leaderboard-server/internal/db/postgresql"
	redis_provider "go-leaderboard-server/internal/db/redis"
	"go-leaderboard-server/internal/utils"
)

type LeaderboardService struct {
	config        *config.Config
	dbprovider    dbprovider.IDbProvider
	cacheprovider cacheprovider.ICacheProvider
}

func NewLeaderboardService(config *config.Config) *LeaderboardService {
	return &LeaderboardService{
		config: config,
	}
}

func (s *LeaderboardService) Initialize(ctx context.Context, clock *utils.IClock) error {
	var (
		err error
	)

	logger.Debug("Leaderboard service initialization")

	switch s.config.Db.Type {
	case config.DBTYPE_INMEMORY:
		s.dbprovider = db_inmemory_provider.NewDbInMemoryProvider()
	case config.DBTYPE_REDIS:
		s.dbprovider = redis_provider.NewRedisProvider()
	case config.DBTYPE_MONGO:
		s.dbprovider = mongo_provider.NewMongoProvider()
	case config.DBTYPE_POSTGRESQL:
		s.dbprovider = postgre_provider.NewPostgreProvider()
	case config.DBTYPE_MYSQL:
		s.dbprovider = mysql_provider.NewMySqlProvider()
	default:
		return errors.New("unknown DB provider type")
	}

	switch s.config.Cache.Type {
	case config.CACHETYPE_SIMPLE:
		s.cacheprovider = cache_simple_provider.NewCacheSimpleProvider(clock)
	default:
		return errors.New("unknown Cache provider type")
	}

	err = s.dbprovider.Initialize(ctx, s.config.Db.Config)
	if err != nil {
		return err
	}

	err = s.cacheprovider.Initialize(ctx, s.config.Cache.Config, s.dbprovider)
	if err != nil {
		return err
	}

	return nil
}

func (s *LeaderboardService) PutUserScore(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	return s.dbprovider.Put(ctx, gameId, userId, userProp)
}

func (s *LeaderboardService) DeleteUserScore(ctx context.Context, gameId string, userId string) error {
	return s.dbprovider.Delete(ctx, gameId, userId)
}

func (s *LeaderboardService) GetUserScore(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	return s.dbprovider.Get(ctx, gameId, userId)
}

func (s *LeaderboardService) GetTop(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	return s.cacheprovider.Top(ctx, gameId, nTop)
}

func (s *LeaderboardService) Shutdown(ctx context.Context) error {
	var (
		err1, err2 error
	)

	logger.Debug("Leaderboard service shutdown")

	if s.dbprovider != nil {
		err1 = s.dbprovider.Shutdown(ctx)
	}

	if s.cacheprovider != nil {
		err2 = s.cacheprovider.Shutdown(ctx)
	}

	err := errors.Join(err1, err2)
	return err
}
