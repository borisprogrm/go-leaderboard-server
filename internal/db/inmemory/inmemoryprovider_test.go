package db_inmemory_provider

import (
	"context"
	dbprovider "go-leaderboard-server/internal/db"
	"go-leaderboard-server/internal/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest() (func() error, *DbInMemoryProvider, error) {
	dbProvider := NewDbInMemoryProvider()
	err := dbProvider.Initialize(context.Background(), &DbInMemoryProviderConfig{
		DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
			IsDebug: true,
		},
	})

	return func() error {
		return dbProvider.Shutdown(context.Background())
	}, dbProvider, err
}

func runTest(t *testing.T, name string, testFunc utils.TestFcn[*DbInMemoryProvider]) {
	utils.RunTest(t, name, setupTest, testFunc)
}

func TestDbInMemoryProvider(t *testing.T) {
	gameId := "game1"
	userId1 := "user1"
	userId2 := "user2"

	userProp1 := dbprovider.UserProperties{
		Name:   "Ted",
		Score:  11,
		Params: "some_payload_1",
	}

	userProp2 := dbprovider.UserProperties{
		Name:  "Jane",
		Score: 56,
	}

	topData := dbprovider.TopData{
		{UserId: userId2, UserProperties: userProp2},
		{UserId: userId1, UserProperties: userProp1},
	}

	runTest(t, "get empty data", func(t *testing.T, dbProvider *DbInMemoryProvider) {
		data, err := dbProvider.Get(context.Background(), gameId, userId1)
		require.NoError(t, err)
		require.Nil(t, data)
	})

	runTest(t, "set, get and delete data", func(t *testing.T, dbProvider *DbInMemoryProvider) {
		var (
			data *dbprovider.UserProperties
			err  error
		)

		err = dbProvider.Put(context.Background(), gameId, userId1, userProp1)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId, userId1)
		require.NoError(t, err)
		require.Equal(t, *data, userProp1)

		var userProp1Mod = userProp1
		userProp1Mod.Score = 33
		err = dbProvider.Put(context.Background(), gameId, userId1, userProp1Mod)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId, userId1)
		require.NoError(t, err)
		require.Equal(t, *data, userProp1Mod)

		err = dbProvider.Delete(context.Background(), gameId, userId1)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId, userId1)
		require.NoError(t, err)
		require.Nil(t, data)
	})

	runTest(t, "get top data", func(t *testing.T, dbProvider *DbInMemoryProvider) {
		var (
			top dbprovider.TopData
			err error
		)

		top, err = dbProvider.Top(context.Background(), gameId, 10)
		require.NoError(t, err)
		require.Equal(t, top, dbprovider.TopData{})

		err = dbProvider.Put(context.Background(), gameId, userId1, userProp1)
		require.NoError(t, err)
		err = dbProvider.Put(context.Background(), gameId, userId2, userProp2)
		require.NoError(t, err)

		top, err = dbProvider.Top(context.Background(), gameId, 10)
		require.NoError(t, err)
		require.Equal(t, top, topData)

		top, err = dbProvider.Top(context.Background(), gameId, 1)
		require.NoError(t, err)
		require.Equal(t, top, topData[:1])
	})

}
