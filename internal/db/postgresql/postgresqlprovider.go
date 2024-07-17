package postgre_provider

import (
	"context"
	"errors"
	"fmt"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var logger = log.GetLogger()

type PostgreProviderConfig struct {
	dbprovider.DBProviderBaseConfig
	ConnStr string
}

type PostgreUserProperties struct {
	Score  dbprovider.UScoreType `db:"score"`
	Name   *string               `db:"name"`
	Params *string               `db:"params"`
}

type PostgreUserData struct {
	UserId string `db:"userId"`
	PostgreUserProperties
}

const DB_TABLE_NAME string = "UserData"

type PostgreProvider struct {
	pool *pgxpool.Pool
}

func NewPostgreProvider() *PostgreProvider {
	return &PostgreProvider{}
}

func toStringOrNull(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (p *PostgreProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	logger.Debug("DB provider initialization")

	conf, ok := config.(*PostgreProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}

	var err error
	pool, err := pgxpool.New(ctx, conf.ConnStr)
	if err != nil {
		return err
	}

	p.pool = pool

	err = pool.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgreProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	_, err := p.pool.Exec(ctx,
		fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT(gameId, userId) DO UPDATE SET
			score = EXCLUDED.score, name = EXCLUDED.name, params = EXCLUDED.params`, DB_TABLE_NAME),
		gameId, userId, userProp.Score, toStringOrNull(userProp.Name), toStringOrNull(userProp.Params),
	)

	return err
}

func (p *PostgreProvider) Delete(ctx context.Context, gameId string, userId string) error {
	_, err := p.pool.Exec(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE gameId = $1 AND userId = $2`, DB_TABLE_NAME),
		gameId, userId,
	)

	return err
}

func (p *PostgreProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	var err error
	rows, err := p.pool.Query(ctx,
		fmt.Sprintf(`SELECT score, name, params FROM %s WHERE gameId = $1 AND userId = $2`, DB_TABLE_NAME),
		gameId, userId,
	)
	if err != nil {
		return nil, err
	}

	uprop, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[PostgreUserProperties])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &dbprovider.UserProperties{
		Score:  uprop.Score,
		Name:   toString(uprop.Name),
		Params: toString(uprop.Params),
	}, nil
}

func (p *PostgreProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	var err error
	rows, err := p.pool.Query(ctx,
		fmt.Sprintf(`SELECT userId as "userId", score, name, params FROM %s
			WHERE gameId = $1 ORDER BY gameId ASC, score DESC LIMIT $2`, DB_TABLE_NAME),
		gameId, nTop,
	)
	if err != nil {
		return dbprovider.TopData{}, err
	}

	var result dbprovider.TopData = make(dbprovider.TopData, 0, nTop)
	top, err := pgx.CollectRows(rows, pgx.RowToStructByName[PostgreUserData])
	if err != nil {
		return dbprovider.TopData{}, err
	}

	for _, udata := range top {
		result = append(result, dbprovider.UserData{
			UserId: udata.UserId,
			UserProperties: dbprovider.UserProperties{
				Score:  udata.Score,
				Name:   toString(udata.Name),
				Params: toString(udata.Params),
			},
		})
	}

	return result, nil
}

func (p *PostgreProvider) Shutdown(ctx context.Context) error {
	if p.pool == nil {
		return nil
	}

	logger.Debug("DB provider shutdown")

	p.pool.Close()

	return nil
}
