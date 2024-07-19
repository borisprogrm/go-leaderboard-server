package redis_provider

import (
	"context"
	dbprovider "go-leaderboard-server/internal/db"
	"go-leaderboard-server/internal/utils"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var dbEndpoint string

func prepareTest(t *testing.T) {
	t.Log("prepare test env")

	ctx := context.Background()
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7.2.3",
			ExposedPorts: []string{"6379"},
			WaitingFor:   wait.ForExposedPort(),
		},
		Started: true,
	})
	require.NoError(t, err, "container should start successfully")

	t.Cleanup(func() {
		t.Log("terminate test env")

		err := dbContainer.Terminate(ctx)
		require.NoError(t, err, "container should be terminated successfully")
	})

	ep, err := dbContainer.Endpoint(ctx, "")
	require.NoError(t, err, "container endpoint should be obtained successfully")

	dbEndpoint = ep
}

func setupTest() (func() error, *RedisProvider, error) {
	dbProvider := NewRedisProvider()
	err := dbProvider.Initialize(context.Background(), &RedisProviderConfig{
		DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
			IsDebug: true,
		},
		Opts: RedisOptions{
			Addr: dbEndpoint,
		},
	})

	return func() error {
		return dbProvider.Shutdown(context.Background())
	}, dbProvider, err
}

func runTest(t *testing.T, name string, testFunc utils.TestFcn[*RedisProvider]) {
	utils.RunTest(t, name, setupTest, testFunc)
}

func TestRedisProvider(t *testing.T) {
	prepareTest(t)

	gameId1 := "game1"
	gameId2 := "game2"
	gameId3 := "game3"
	userId1 := "user1"
	userId2 := "user2"

	userProp1 := dbprovider.UserProperties{
		Name:   "Ted",
		Score:  15,
		Params: "some_payload_1",
	}

	userProp2 := dbprovider.UserProperties{
		Score: 62,
	}

	topData := dbprovider.TopData{
		{UserId: userId2, UserProperties: userProp2},
		{UserId: userId1, UserProperties: userProp1},
	}

	runTest(t, "get empty data", func(t *testing.T, dbProvider *RedisProvider) {
		data, err := dbProvider.Get(context.Background(), gameId1, userId1)
		require.NoError(t, err)
		require.Nil(t, data)
	})

	runTest(t, "set, get and delete data", func(t *testing.T, dbProvider *RedisProvider) {
		var (
			data *dbprovider.UserProperties
			err  error
		)

		err = dbProvider.Put(context.Background(), gameId2, userId1, userProp1)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		require.Equal(t, userProp1, *data)

		var userProp1Mod = userProp1
		userProp1Mod.Score = 33
		err = dbProvider.Put(context.Background(), gameId2, userId1, userProp1Mod)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		require.Equal(t, userProp1Mod, *data)

		err = dbProvider.Delete(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		require.Nil(t, data)
	})

	runTest(t, "get top data", func(t *testing.T, dbProvider *RedisProvider) {
		var (
			top dbprovider.TopData
			err error
		)

		top, err = dbProvider.Top(context.Background(), gameId3, 10)
		require.NoError(t, err)
		require.Equal(t, dbprovider.TopData{}, top)

		err = dbProvider.Put(context.Background(), gameId3, userId1, userProp1)
		require.NoError(t, err)
		err = dbProvider.Put(context.Background(), gameId3, userId2, userProp2)
		require.NoError(t, err)

		top, err = dbProvider.Top(context.Background(), gameId3, 10)
		require.NoError(t, err)
		require.Equal(t, topData, top)

		top, err = dbProvider.Top(context.Background(), gameId3, 1)
		require.NoError(t, err)
		require.Equal(t, topData[:1], top)
	})

}
