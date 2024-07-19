package dynamo_provider

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"sort"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var logger = log.GetLogger()

type DynamoProviderConfig struct {
	dbprovider.DBProviderBaseConfig
	Region          string // The region to send requests to
	Endpoint        string // Endpoint
	AccessKeyId     string // AWS Access key ID
	SecretAccessKey string // AWS Secret Access Key
	NShards         uint32 // Number of custom shards used to avoid "hot" partition problem (0-100)
}

const DBTABLE_NAME string = "Leaderboard"
const DBTABLE_INDEX_NAME string = "ScoreIndex"

type DynamoProvider struct {
	db      *dynamodb.Client
	nShards uint32
}

func NewDynamoProvider() *DynamoProvider {
	return &DynamoProvider{}
}

func getHashKey(gameId string, userId string, nShards uint32) string {
	h := sha1.New()
	h.Write([]byte(userId))
	hash := h.Sum(nil)

	var shard uint32 = 0
	if nShards > 1 {
		shard = binary.BigEndian.Uint32(hash[len(hash)-4:]) % nShards
	}

	key := fmt.Sprintf("%s:%d", gameId, shard)
	return key
}

func (p *DynamoProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	logger.Debug("DB provider initialization")

	conf, ok := config.(*DynamoProviderConfig)
	if !ok {
		return errors.New("wrong config")
	}
	if conf.NShards > 100 {
		return errors.New("wrong config: NShards")
	}
	p.nShards = conf.NShards

	awsConf := aws.Config{
		Region:       conf.Region,
		BaseEndpoint: &conf.Endpoint,
		Credentials:  credentials.NewStaticCredentialsProvider(conf.AccessKeyId, conf.SecretAccessKey, ""),
	}

	client := dynamodb.NewFromConfig(awsConf)
	p.db = client

	return nil
}

func (p *DynamoProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	item := map[string]types.AttributeValue{
		"gId": &types.AttributeValueMemberS{Value: getHashKey(gameId, userId, p.nShards)},
		"uId": &types.AttributeValueMemberS{Value: userId},
		"sc":  &types.AttributeValueMemberN{Value: strconv.FormatFloat(float64(userProp.Score), 'f', -1, 64)},
	}

	if userProp.Name != "" {
		item["nm"] = &types.AttributeValueMemberS{Value: userProp.Name}
	}
	if userProp.Params != "" {
		item["pl"] = &types.AttributeValueMemberS{Value: userProp.Params}
	}

	_, err := p.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(DBTABLE_NAME),
		Item:      item,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *DynamoProvider) Delete(ctx context.Context, gameId string, userId string) error {
	key := map[string]types.AttributeValue{
		"gId": &types.AttributeValueMemberS{Value: getHashKey(gameId, userId, p.nShards)},
		"uId": &types.AttributeValueMemberS{Value: userId},
	}

	_, err := p.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(DBTABLE_NAME),
		Key:       key,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *DynamoProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	key := map[string]types.AttributeValue{
		"gId": &types.AttributeValueMemberS{Value: getHashKey(gameId, userId, p.nShards)},
		"uId": &types.AttributeValueMemberS{Value: userId},
	}

	result, err := p.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(DBTABLE_NAME),
		Key:            key,
		ConsistentRead: aws.Bool(true), // this can actually be skipped if necessary (lower cost)
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var item dbprovider.UserProperties
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (p *DynamoProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	type Result struct {
		result *dynamodb.QueryOutput
		err    error
	}

	N := max(p.nShards, 1)
	resChan := make(chan Result, N)

	var wg sync.WaitGroup
	QueryAsync := func(idx uint32) {
		defer wg.Done()
		result, err := p.db.Query(ctx, &dynamodb.QueryInput{
			TableName: aws.String(DBTABLE_NAME),
			IndexName: aws.String(DBTABLE_INDEX_NAME),
			KeyConditions: map[string]types.Condition{
				"gId": {
					ComparisonOperator: types.ComparisonOperatorEq,
					AttributeValueList: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: fmt.Sprintf("%s:%d", gameId, idx)},
					},
				},
			},
			ScanIndexForward: aws.Bool(false),
			Limit:            aws.Int32(int32(nTop)),
		})
		resChan <- Result{result, err}
	}

	for i := uint32(0); i < N; i++ {
		wg.Add(1)
		go QueryAsync(i)
	}

	wg.Wait()

	results := make([]*dynamodb.QueryOutput, N)
	var nitems = 0
	for i := range results {
		res := <-resChan
		if res.err != nil {
			return dbprovider.TopData{}, res.err
		}
		results[i] = res.result
		nitems += len(res.result.Items)
	}

	uscores := make(dbprovider.TopData, 0, nitems)
	for _, result := range results {
		for _, item := range result.Items {
			var udata dbprovider.UserData
			err := attributevalue.UnmarshalMap(item, &udata)
			if err != nil {
				return dbprovider.TopData{}, err
			}
			uscores = append(uscores, udata)
		}
	}

	if N > 1 {
		sort.Slice(uscores, func(i, j int) bool {
			return uscores[j].UserProperties.Score < uscores[i].UserProperties.Score
		})
	}

	top := uscores[:min(len(uscores), int(nTop))]
	return top, nil
}

func (p *DynamoProvider) Shutdown(ctx context.Context) error {
	if p.db == nil {
		return nil
	}

	logger.Debug("DB provider shutdown")

	/* do nothing */

	return nil
}
