package mongo_provider

import (
	"context"
	"errors"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var logger = log.GetLogger()

type MongoProviderConfig struct {
	dbprovider.DBProviderBaseConfig
	Uri     string
	Options *options.ClientOptions
}

type MongoUserID struct {
	UserId string `bson:"uId"`
}

type MongoUserData struct {
	MongoUserID               `bson:"_id"`
	dbprovider.UserProperties `bson:",inline"`
}

const DB_NAME string = "GoLeaderboard"
const DB_COLLECTION_NAME string = "UserData"

type MongoProvider struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoProvider() *MongoProvider {
	return &MongoProvider{}
}

func (p *MongoProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	logger.Debug("DB provider initialization")

	conf, ok := config.(*MongoProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}

	dbclient, err := mongo.Connect(ctx, conf.Options.ApplyURI(conf.Uri))
	if err != nil {
		return err
	}

	p.client = dbclient
	err = p.client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	p.collection = p.client.Database(DB_NAME).Collection(DB_COLLECTION_NAME)

	return nil
}

func (p *MongoProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "gId", Value: gameId}, {Key: "uId", Value: userId}}}}
	opts := options.Replace().SetUpsert(true)
	_, err := p.collection.ReplaceOne(ctx, filter, userProp, opts)
	if err != nil {
		return err
	}

	return nil
}

func (p *MongoProvider) Delete(ctx context.Context, gameId string, userId string) error {
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "gId", Value: gameId}, {Key: "uId", Value: userId}}}}
	_, err := p.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (p *MongoProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	var result dbprovider.UserProperties
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "gId", Value: gameId}, {Key: "uId", Value: userId}}}}
	err := p.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

func (p *MongoProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	filter := bson.D{{Key: "_id.gId", Value: gameId}}
	opts := options.Find().SetHint("ScoreIndex").SetSort(bson.D{{Key: "sc", Value: -1}}).SetLimit(int64(nTop))
	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		return dbprovider.TopData{}, err
	}

	var result dbprovider.TopData = make(dbprovider.TopData, 0, nTop)

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var mres MongoUserData
		err := cursor.Decode(&mres)
		if err != nil {
			return dbprovider.TopData{}, err
		}

		result = append(result, dbprovider.UserData{
			UserId:         mres.UserId,
			UserProperties: mres.UserProperties,
		})

	}
	err = cursor.Err()
	if err != nil {
		return dbprovider.TopData{}, err
	}

	return result, nil
}

func (p *MongoProvider) Shutdown(ctx context.Context) error {
	if p.client == nil {
		return nil
	}

	logger.Debug("DB provider shutdown")

	err := p.client.Disconnect(ctx)

	return err
}
