package redis_provider

import (
	"context"
	"errors"
	"fmt"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"

	"github.com/redis/go-redis/v9"
)

var logger = log.GetLogger()

type RedisOptions redis.Options

type RedisProviderConfig struct {
	dbprovider.DBProviderBaseConfig
	Opts RedisOptions
}

type RedisProvider struct {
	rdb *redis.Client
}

func NewRedisProvider() *RedisProvider {
	return &RedisProvider{}
}

func getUserKey(gameId string, userId string) string {
	return fmt.Sprintf("%s:%s", gameId, userId)
}

func (p *RedisProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	logger.Debug("DB provider initialization")

	conf, ok := config.(*RedisProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}

	opts := redis.Options(conf.Opts)
	rdb := redis.NewClient(&opts)
	p.rdb = rdb

	err := p.rdb.Ping(ctx).Err()
	if err != nil {
		return err
	}

	return nil
}

func (p *RedisProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	var err error

	if userProp.Name != "" || userProp.Params != "" {
		hval := make(map[string]string)
		if userProp.Name != "" {
			hval["nm"] = userProp.Name
		}
		if userProp.Params != "" {
			hval["pl"] = userProp.Params
		}

		err = p.rdb.HSet(ctx, getUserKey(gameId, userId), hval).Err()
		if err != nil {
			return err
		}
	}

	err = p.rdb.ZAdd(ctx, gameId, redis.Z{
		Score:  float64(userProp.Score),
		Member: userId,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func (p *RedisProvider) Delete(ctx context.Context, gameId string, userId string) error {
	var err error

	err = p.rdb.ZRem(ctx, gameId, userId).Err()
	if err != nil {
		return err
	}

	err = p.rdb.Del(ctx, getUserKey(gameId, userId)).Err()
	if err != nil {
		return err
	}

	return nil
}

func (p *RedisProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	type Result[T any] struct {
		value T
		err   error
	}
	chanScore := make(chan Result[float64])
	chanHash := make(chan Result[map[string]string])

	go func() {
		score, err := p.rdb.ZScore(ctx, gameId, userId).Result()
		chanScore <- Result[float64]{score, err}
	}()

	go func() {
		hval, err := p.rdb.HGetAll(ctx, getUserKey(gameId, userId)).Result()
		chanHash <- Result[map[string]string]{hval, err}
	}()

	resScore := <-chanScore
	if resScore.err != nil {
		if resScore.err == redis.Nil {
			return nil, nil
		}
		return nil, resScore.err
	}

	resHash := <-chanHash
	if resHash.err != nil {
		return nil, resHash.err
	}

	return &dbprovider.UserProperties{
		Score:  dbprovider.UScoreType(resScore.value),
		Name:   resHash.value["nm"],
		Params: resHash.value["pl"],
	}, nil
}

func (p *RedisProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	var err error

	topData, err := p.rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:   gameId,
		Start: 0,
		Stop:  int64(nTop) - 1,
		Rev:   true,
	}).Result()
	if err != nil {
		return dbprovider.TopData{}, err
	}

	N := len(topData)
	if N < 1 {
		return dbprovider.TopData{}, nil
	}

	pipe := p.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, N)
	for i, key := range topData {
		cmds[i] = pipe.HGetAll(ctx, getUserKey(gameId, key.Member.(string)))
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return dbprovider.TopData{}, err
	}

	var top dbprovider.TopData = make(dbprovider.TopData, 0, N)
	for i, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			return dbprovider.TopData{}, nil
		}
		top = append(top, dbprovider.UserData{
			UserId: topData[i].Member.(string),
			UserProperties: dbprovider.UserProperties{
				Score:  dbprovider.UScoreType(topData[i].Score),
				Name:   result["nm"],
				Params: result["pl"],
			},
		})
	}

	return top, nil
}

func (p *RedisProvider) Shutdown(ctx context.Context) error {
	if p.rdb == nil {
		return nil
	}

	logger.Debug("DB provider shutdown")

	err := p.rdb.Close()

	return err
}
