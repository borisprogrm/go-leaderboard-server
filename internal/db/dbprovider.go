package dbprovider

import "context"

type UScoreType float64

type UserProperties struct {
	Score  UScoreType `json:"score" bson:"sc"`
	Name   string     `json:"name,omitempty" bson:"nm,omitempty"`
	Params string     `json:"params,omitempty" bson:"pl,omitempty"`
}

type UserData struct {
	UserId         string `json:"userId" bson:"uId" binding:"required"`
	UserProperties `bson:",inline"`
}

type TopData []UserData

type DBProviderBaseConfig struct {
	IsDebug bool // Debug flag
}

func (c *DBProviderBaseConfig) GetBaseConfig() *DBProviderBaseConfig {
	return c
}

type IDBProviderConfig interface {
	GetBaseConfig() *DBProviderBaseConfig
}

type IDbProvider interface {
	Initialize(ctx context.Context, config IDBProviderConfig) error
	Put(ctx context.Context, gameId string, userId string, userProp UserProperties) error
	Delete(ctx context.Context, gameId string, userId string) error
	Get(ctx context.Context, gameId string, userId string) (*UserProperties, error)
	Top(ctx context.Context, gameId string, nTop uint32) (TopData, error)
	Shutdown(ctx context.Context) error
}
