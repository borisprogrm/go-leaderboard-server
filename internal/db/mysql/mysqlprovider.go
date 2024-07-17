package mysql_provider

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"time"

	"github.com/georgysavva/scany/sqlscan"
	_ "github.com/go-sql-driver/mysql"
)

var logger = log.GetLogger()

type MySqlProviderConfig struct {
	dbprovider.DBProviderBaseConfig
	ConnStr         string // Connection string
	MaxOpenConns    uint32 // Maximum number of open connections to the database
	MaxIdleConns    uint32 // Maximum number of connections in the idle connection pool
	ConnMaxLifetime uint32 // Maximum amount of time a connection may be reused (ms)
	ConnMaxIdleTime uint32 // Maximum amount of time a connection may be idle (ms)
}

type MySqlUserProperties struct {
	Score  dbprovider.UScoreType `db:"score"`
	Name   *string               `db:"name"`
	Params *string               `db:"params"`
}

type MySqlUserData struct {
	UserId string `db:"userId"`
	MySqlUserProperties
}

const DB_TABLE_NAME string = "UserData"

type MySqlProvider struct {
	db *sql.DB
}

func NewMySqlProvider() *MySqlProvider {
	return &MySqlProvider{}
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

func (p *MySqlProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	logger.Debug("DB provider initialization")

	conf, ok := config.(*MySqlProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}

	var err error
	db, err := sql.Open("mysql", conf.ConnStr)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(int(conf.MaxOpenConns))
	db.SetMaxIdleConns(int(conf.MaxIdleConns))
	db.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Millisecond)
	db.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Millisecond)

	p.db = db

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p *MySqlProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	_, err := p.db.ExecContext(ctx,
		fmt.Sprintf(`INSERT INTO %s VALUES (?, ?, ?, ?, ?) AS new
			ON DUPLICATE KEY UPDATE
			score = new.score, name = new.name, params = new.params`, DB_TABLE_NAME),
		gameId, userId, userProp.Score, toStringOrNull(userProp.Name), toStringOrNull(userProp.Params),
	)

	return err
}

func (p *MySqlProvider) Delete(ctx context.Context, gameId string, userId string) error {
	_, err := p.db.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE gameId = ? AND userId = ?`, DB_TABLE_NAME),
		gameId, userId,
	)

	return err
}

func (p *MySqlProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	var err error
	rows, err := p.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT score, name, params FROM %s WHERE gameId = ? AND userId = ?`, DB_TABLE_NAME),
		gameId, userId,
	)
	if err != nil {
		return nil, err
	}

	var uprop MySqlUserProperties
	err = sqlscan.ScanOne(&uprop, rows)
	if err != nil {
		if sqlscan.NotFound(err) {
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

func (p *MySqlProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	var err error
	rows, err := p.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT userId as "userId", score, name, params FROM %s
			WHERE gameId = ? ORDER BY gameId ASC, score DESC LIMIT ?`, DB_TABLE_NAME),
		gameId, nTop,
	)
	if err != nil {
		return dbprovider.TopData{}, err
	}

	var top []MySqlUserData
	err = sqlscan.ScanAll(&top, rows)
	if err != nil {
		return dbprovider.TopData{}, err
	}

	var result dbprovider.TopData = make(dbprovider.TopData, 0, nTop)
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

func (p *MySqlProvider) Shutdown(ctx context.Context) error {
	if p.db == nil {
		return nil
	}

	logger.Debug("DB provider shutdown")

	p.db.Close()

	return nil
}
